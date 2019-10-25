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
	//assert that the interface type of err is CoulsonError
	if *err != nil {
		r.rs.ReturnCode = common.ReturnCodeFailed
		r.rs.ReturnMsg = (*err).Error()
		if ce, ok := errors.Cause(*err).(CoulsonError); ok {
			log.Slogger.Errorf("encounter an error:%+v, the kv is: %s", *err, ce.Kv())
		} else {
			log.Slogger.Errorf("encounter an error:%+v.", *err)
		}
	}
	//write the result to chanel
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
		if err != nil {
			r.rs.AppendFailedStep(stepNameStart, err)
			return
		}
		r.rs.AppendSuccessStep(stepNameStart)
	case common.OperateSHU:
		err = r.shutdown()
		if err != nil {
			r.rs.AppendFailedStep(stepNameStop, err)
			return
		}
		r.rs.AppendSuccessStep(stepNameStop)
	case common.OperateRES:
		err = r.shutdown()
		if err != nil {
			r.rs.AppendFailedStep(stepNameStop, err)
			return
		}

		err = r.start()
		if err != nil {
			r.rs.AppendFailedStep(stepNameStart, err)
			return
		}
		r.rs.AppendSuccessStep(stepNameRestart)
	}
}

//start the program
func (r *Rss) start() error {
	//run the startup command
	err := r.runCMD(r.StartCMD)
	if err != nil {
		return err
	}
	//achieve the path of register script
	cmdMetaScript, err := r.getMetaScriptPath()
	if err != nil {
		return err
	}
	//Run the registration script
	err = r.runCMD(cmdMetaScript)
	if err != nil {
		return err
	}

	return nil
}

//shutdown the program
func (r *Rss) shutdown() error {
	err := r.runCMD(r.StopCMD)
	if err != nil {
		return err
	}
	return nil
}

//get the meta script path which will be executed
//first, walk the base dir of the service, and just traverse the first level directory
//If you have a "bin" directory, use the script under there, otherwise use script under base directory
func (r *Rss) getMetaScriptPath() (string, error) {
	var cmdMetaScript = filepath.Join(r.Dir, common.RegisterScript)
	toFindDir := "bin"
	err := afis.WalkOnce(r.Dir, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			return err
		}
		log.Slogger.Debugf("walk path: %s, info name: %s", path, info.Name())

		if info.IsDir() && info.Name() == toFindDir {
			cmdMetaScript = filepath.Join(path, common.RegisterScript)
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
	log.Slogger.Debugf("the meta script is: %s", cmdMetaScript)

	if !afis.IsFile(cmdMetaScript) {
		return "", errors.WithStack(NewPathError(cmdMetaScript, "no meta script"))
	}
	return cmdMetaScript, nil
}

//run the command
func (r *Rss) runCMD(cmdString string) error {

	// changes the current working directory to bin directory
	// for afis programe which can't be executed witch absolute path
	binPath := filepath.Join(r.Dir, "/bin")
	err := os.Chdir(binPath)
	if err != nil {
		return errors.WithStack(NewPathError(binPath, err.Error()))

	}
	_, err = exec.LookPath(cmdString)
	if err != nil {
		return errors.WithStack(NewPathError(cmdString, err.Error()))
	}

	cmdSlice := strings.Fields(strings.TrimSpace(cmdString))

	log.Slogger.Debugf("The cmdSlice is:%+v", cmdSlice)

	cmd := exec.Command(cmdSlice[0], cmdSlice[1:]...)

	userInfo, err := r.getUserInfo()
	if err != nil {
		return err
	}

	//switch os user
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(userInfo.Uid), Gid: uint32(userInfo.Gid)}
	//append some environment variables
	cmd.Env = append(os.Environ(), "HOME="+userInfo.HomeDir, "USER="+userInfo.Username, "SHELL="+"/bin/bash")

	out, err := cmd.CombinedOutput()

	log.Slogger.Debugf("stdout:\n%s", string(out))
	if err != nil {
		return errors.WithStack(NewCmdError(cmdString, err.Error()))
	}

	return nil
}

//achieve  userinfo
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
