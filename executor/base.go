package executor

import (
	"fmt"
	"github.com/glory-cd/agent/common"
	"github.com/glory-cd/utils/afis"
	"github.com/glory-cd/utils/log"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

type Driver struct {
	*Task
	*Service
}

// Backup service dir to temporary file and upload to file server
func (d *Driver) BackupService(tmpdst, uploadpath string) error {
	// Create if the tmpdst directory does not exist
	if !afis.IsExists(filepath.Dir(tmpdst)) {
		err := os.MkdirAll(filepath.Dir(tmpdst), 0755)
		if err != nil {
			return err
		}
	}
	// Compressed files
	src := d.Dir
	err := afis.Zipit(src, tmpdst, "*.log")
	if err != nil {
		return errors.WithStack(err)
	}
	// Upload to the file server
	fileServer := common.Config().FileServer
	err = Upload(
		fileServer,
		tmpdst,
		uploadpath)

	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (d *Driver) DeleteService() error {
	//Return error if the file that you want to delete belongs to a different user
	if !afis.CheckFileOwner(d.Dir, d.OsUser) {
		return errors.WithStack(NewFileOwnerError(d.Dir, d.OsUser, "file and owner does not match"))
	}
	// Delete the whole service directory
	err := os.RemoveAll(d.Dir)
	if err != nil {

		return errors.WithStack(err)
	}

	return nil
}

// Read the PathFile file and get the backup path in the FileServer
func (d *Driver) ReadServiceVerion() (string, error) {
	versionFile := filepath.Join(d.Dir, common.PathFile)
	path, err := ioutil.ReadFile(versionFile)
	if err != nil {
		return "", errors.WithStack(err)
	}

	return strings.TrimSpace(string(path)), nil
}

// Get the execution path of command
func (d *Driver) GetBinPath(cmd string) (string, error) {
	var cmdpath string

	path := os.Getenv("PATH")

	pathSlice := strings.Split(path, string(os.PathListSeparator))

	for _, p := range pathSlice {
		fullcmd := filepath.Join(p, cmd)
		if afis.IsFile(fullcmd) && afis.IsExecutable(fullcmd) {
			cmdpath = fullcmd
			break
		}
	}

	if cmdpath == "" {
		return "", fmt.Errorf("command not found: %s", cmd)
	}
	return cmdpath, nil
}

// Download code from FileServer
func (d *Driver) GetCode() (string, error) {

	//download code from url
	fileServer := common.Config().FileServer
	dir, err := Get(fileServer, d.RemoteCode)
	if err != nil {
		return "", err
	}

	log.Slogger.Infof("download code to %s", dir)
	return dir, nil
}

// Get the meta script path which will be executed
func (d *Driver) GetMetaScript() (string, error) {

	cmdBin, err := d.getScriptBinPath()

	if err != nil {
		return "", err
	}

	cmdMetaScript := filepath.Join(cmdBin, common.RegisterScript)

	log.Slogger.Debugf("the meta script is: %s", cmdMetaScript)

	if !afis.IsFile(cmdMetaScript) {
		return "", errors.WithStack(NewPathError(cmdMetaScript, "no meta script"))
	}
	return cmdMetaScript, nil
}

// Get the  script bin path which Contains executable scripts
// First, walk the base dir of the service, and just traverse the first level directory
// If you have a "bin" directory, then use it, otherwise use base directory
func (d *Driver) getScriptBinPath() (string, error) {
	var cmdScriptPath = d.Dir
	toFindDir := "bin"
	err := afis.WalkOnce(d.Dir, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			return err
		}

		if info.IsDir() && info.Name() == toFindDir {
			cmdScriptPath = path
			return io.EOF
		}

		return nil

	})

	if errors.Cause(err) == io.EOF {
		err = nil
	}

	if err != nil {
		return "", err
	}

	log.Slogger.Debugf("the script bin is: %s", cmdScriptPath)

	return cmdScriptPath, nil
}

// Run the command
// If cmdString contains a slash, it is tried directly and the PATH is not consulted.
// or find the command from PATH
func (d *Driver) RunCMD(cmdString string) error {

	// Changes the current working directory to bin directory
	cmdBin, err := d.getScriptBinPath()

	if err != nil {
		return err
	}

	err = os.Chdir(cmdBin)
	if err != nil {
		return errors.WithStack(NewPathError(cmdBin, err.Error()))
	}

	cmdSlice := strings.Fields(strings.TrimSpace(cmdString))

	log.Slogger.Debugf("The cmdSlice is:%+v", cmdSlice)

	if len(cmdSlice) == 0 {
		return errors.New("index out of range [0] with length 0")
	}
	// search for an executable named file in the
	// directories named by the PATH environment variable.
	// If file contains a slash, it is tried directly and the PATH is not consulted.
	// The result may be an absolute path or a path relative to the current directory.
	executableFile, err := exec.LookPath(cmdSlice[0])
	if err != nil {
		return errors.WithStack(NewPathError(cmdSlice[0], err.Error()))
	}

	cmd := exec.Command(executableFile, cmdSlice[1:]...)

	userInfo, err := d.getUserInfo()
	if err != nil {
		return err
	}

	// Switch os user
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(userInfo.Uid), Gid: uint32(userInfo.Gid)}
	// Append some environment variables
	cmd.Env = append(os.Environ(), "HOME="+userInfo.HomeDir, "USER="+userInfo.Username)

	out, err := cmd.CombinedOutput()

	log.Slogger.Debugf("stdout:\n%s", string(out))
	if err != nil {
		return errors.WithStack(NewCmdError(cmdString, err.Error()))
	}

	return nil
}


// Achieve  userinfo
func (d *Driver) getUserInfo() (userInfo *User, err error) {
	suser, err := user.Lookup(d.OsUser)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	uid, err := strconv.Atoi(suser.Uid)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	gid, err := strconv.Atoi(suser.Gid)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	userInfo = new(User)
	userInfo.Uid = uid
	userInfo.Gid = gid
	userInfo.Name = suser.Name
	userInfo.Username = suser.Username
	userInfo.HomeDir = suser.HomeDir

	return
}