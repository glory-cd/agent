/*
@Time : 19-5-6 下午1:52
@Author : liupeng
@File : rss.go
*/

package executor

import (
	"github.com/glory-cd/agent/common"
	"github.com/glory-cd/utils/afis"
	"github.com/glory-cd/utils/log"
	"github.com/pkg/errors"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

type Rss struct {
	driver
	rs Result
}

func (r *Rss) deferHandleFunc(err *error, out chan<- Result) {
	// Assert that the interface type of err is CoulsonError
	if *err != nil {
		r.rs.ReturnCode = common.ReturnCodeFailed
		r.rs.ReturnMsg = (*err).Error()
		if ce, ok := errors.Cause(*err).(CoulsonError); ok {
			log.Slogger.Errorf("encounter an error:%+v, the kv is: %s", *err, ce.Kv())
		} else {
			log.Slogger.Errorf("encounter an error:%+v.", *err)
		}
	}
	// write the result to chanel
	out <- r.rs
	log.Slogger.Infof("Exit goroutine.")
}

func (r *Rss) Exec(out chan<- Result) {

	var operateString string

	switch r.OP {
	case common.OperateSTA:
		operateString = "START"
	case common.OperateSHU:
		operateString = "STOP"
	case common.OperateRES:
		operateString = "RESTART"
	}
	log.Slogger.Infof("Begin to [%s] service：%s,%s", operateString, r.ServiceID, r.Dir)

	var err error
	defer r.deferHandleFunc(&err, out)

	switch r.OP {
	case common.OperateSTA:
		err = r.start()
	case common.OperateSHU:
		err = r.shutdown()
	case common.OperateRES:
		err = r.restart()
	}
}


// Start the program
func (r *Rss) start() error {
	var err error
	defer func() {
		if err != nil {
			r.rs.AppendFailedStep(stepNameStart, err)
		} else {
			r.rs.AppendSuccessStep(stepNameStart)
		}
	}()
	// Run the startup command
	err = r.runCMD(r.StartCMD)
	if err != nil {
		return err
	}
	// Achieve the path of register script
	cmdMetaScript, err := r.getMetaScript()
	if err != nil {
		return err
	}
	// Run the registration script
	err = r.runCMD(cmdMetaScript)
	if err != nil {
		return err
	}

	return nil
}

// Shutdown the program
func (r *Rss) shutdown() error {
	var err error
	defer func() {
		if err != nil {
			r.rs.AppendFailedStep(stepNameStop, err)
		} else {
			r.rs.AppendSuccessStep(stepNameStop)
		}
	}()

	err = r.runCMD(r.StopCMD)
	if err != nil {
		return err
	}
	return nil
}

// Restart the program
func (r *Rss) restart() error{
	err := r.shutdown()
	if err != nil {
		return err
	}
	err = r.start()
	if err != nil {
		return err
	}

	return nil
}

// Get the meta script path which will be executed
func (r *Rss) getMetaScript() (string, error) {

	cmdBin, err := r.getScriptBinPath()

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
func (r *Rss) getScriptBinPath() (string, error) {
	var cmdScriptPath = r.Dir
	toFindDir := "bin"
	err := afis.WalkOnce(r.Dir, func(path string, info os.FileInfo, err error) error {

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
func (r *Rss) runCMD(cmdString string) error {

	// Changes the current working directory to bin directory
	cmdBin, err := r.getScriptBinPath()

	if err != nil {
		return err
	}

	err = os.Chdir(cmdBin)
	if err != nil {
		return errors.WithStack(NewPathError(cmdBin, err.Error()))
	}

	cmdSlice := strings.Fields(strings.TrimSpace(cmdString))

	log.Slogger.Debugf("The cmdSlice is:%+v", cmdSlice)
	// search for an executable named file in the
	// directories named by the PATH environment variable.
	// If file contains a slash, it is tried directly and the PATH is not consulted.
	// The result may be an absolute path or a path relative to the current directory.
	executableFile, err := exec.LookPath(cmdSlice[0])
	if err != nil {
		return errors.WithStack(NewPathError(cmdSlice[0], err.Error()))
	}

	cmd := exec.Command(executableFile, cmdSlice[1:]...)

	userInfo, err := r.getUserInfo()
	if err != nil {
		return err
	}

	// Switch os user
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(userInfo.Uid), Gid: uint32(userInfo.Gid)}
	// Append some environment variables
	cmd.Env = append(os.Environ(), "HOME="+userInfo.HomeDir, "USER="+userInfo.Username, "SHELL="+"/bin/bash")

	out, err := cmd.CombinedOutput()

	log.Slogger.Debugf("stdout:\n%s", string(out))
	if err != nil {
		return errors.WithStack(NewCmdError(cmdString, err.Error()))
	}

	return nil
}

// Achieve  userinfo
func (r *Rss) getUserInfo() (userInfo *User, err error) {
	suser, err := user.Lookup(r.OsUser)
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
