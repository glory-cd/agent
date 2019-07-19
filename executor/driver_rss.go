/*
@Time : 19-5-6 下午1:52
@Author : liupeng
@File : rss.go
*/

package executor

import (
	"agent/common"
	"bytes"
	"github.com/pkg/errors"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
	"utils/log"
)

type Rss struct {
	*Task
	*Service
	rs Result
}

//构造ResultService
func (r *Rss) constructRS(rcode common.ExecuteReturnCode, stepname, errstr string) {
	r.rs.AppendResultStep(
		1,
		stepname,
		rcode,
		errstr,
		time.Now().UnixNano(),
	)
}

func (r *Rss) deferHandleFunc(err *error, out chan<- Result) {
	//断言err的接口类型为CoulsonError
	if *err != nil {
		r.rs.ReturnCode = common.ReturnCode_FAILED
		r.rs.ReturnMsg = (*err).Error()
		if ce, ok := errors.Cause(*err).(CoulsonError); ok {
			log.Slogger.Errorf("encounter an error:%+v, the kv is: %s", *err, ce.Kv())
		} else {
			log.Slogger.Errorf("encounter an error:%+v.", *err)
		}
	}
	//结果写入chanel
	out <- r.rs
	log.Slogger.Infof("退出goroutine.")
}

func (r *Rss) Exec(out chan<- Result) {

	var operateString string

	switch r.OP {
	case common.Operate_STA:
		operateString = "START"
	case common.Operate_SHU:
		operateString = "STOP"
	case common.Operate_RES:
		operateString = "RESTART"
	}

	log.Slogger.Infof("开始[%s]服务：%s,%s", operateString, r.ServiceID, r.Name)
	var err error
	defer r.deferHandleFunc(&err, out)

	switch r.OP {
	case common.Operate_STA:
		err = r.start()
		if err != nil {
			r.constructRS(common.ReturnCode_FAILED, operateString, err.Error())
			return
		}
		r.constructRS(common.ReturnCode_SUCCESS, operateString, common.ReturnOKMsg)
	case common.Operate_SHU:
		err = r.shutdown()
		if err != nil {
			r.constructRS(common.ReturnCode_FAILED, operateString, err.Error())
			return
		}
		r.constructRS(common.ReturnCode_SUCCESS, operateString, common.ReturnOKMsg)
	case common.Operate_RES:
		err = r.shutdown()
		if err != nil {
			r.constructRS(common.ReturnCode_FAILED, "STOP", err.Error())
			return
		}

		err = r.start()
		if err != nil {
			r.constructRS(common.ReturnCode_FAILED, "START", err.Error())
			return
		}
		r.constructRS(common.ReturnCode_SUCCESS, operateString, common.ReturnOKMsg)
	}
}

//启动程序
func (r *Rss) start() error {
	err := r.runCMD(r.StartCMD)
	if err != nil {
		return err
	}
	return nil
}

//关闭程序
func (r *Rss) shutdown() error {
	err := r.runCMD(r.StopCMD)
	if err != nil {
		return err
	}
	return nil
}

//执行CMD
func (r *Rss) runCMD(cmdString string) error {
	//构造执行文件的bin路径
	binPath := filepath.Join(r.Dir, "/bin")
	err := os.Chdir(binPath)
	if err != nil {
		return errors.WithStack(NewPathError(binPath, err.Error()))

	}
	//检查执行路径是否存在
	_, err = exec.LookPath(cmdString)
	if err != nil {
		return errors.WithStack(NewPathError(cmdString, err.Error()))
	}
	cmd := exec.Command(cmdString)
	//切换用户
	suser, _ := user.Lookup(r.OsUser)
	uid, _ := strconv.Atoi(suser.Uid)
	gid, _ := strconv.Atoi(suser.Gid)
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}
	//改变Env
	cmd.Env = append(os.Environ(), "HOME="+suser.HomeDir, "USER="+suser.Username)
	//处理stdout和stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	//执行
	err = cmd.Run()
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	if err != nil {
		return errors.WithStack(NewCmdError(cmdString, errStr))
	}
	log.Slogger.Infof("stdout:\n%s", outStr)
	return nil
}
