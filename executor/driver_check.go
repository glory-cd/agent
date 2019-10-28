/*
@Time : 19-5-6 下午1:52
@Author : liupeng
@File : check.go
*/

package executor

import (
	"github.com/glory-cd/agent/common"
	"github.com/glory-cd/utils/afis"
	"github.com/glory-cd/utils/log"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/process"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type Check struct {
	driver
	rs Result
}

func (c *Check) Exec(out chan<- Result) {
	log.Slogger.Infof("Begin to [CHECK] service：%s,%s", c.ServiceID, c.Dir)
	var err error
	defer func() {

		// Assert that the interface type of err is CoulsonError
		if err != nil {
			c.rs.ReturnCode = common.ReturnCodeFailed
			c.rs.ReturnMsg = err.Error()
			log.Slogger.Debugf("Result:%+v", c.rs)

			if ce, ok := errors.Cause(err).(CoulsonError); ok {
				log.Slogger.Errorf("encounter an error:%+v, the kv is: %s", err, ce.Kv())
			} else {
				log.Slogger.Errorf("encounter an error:%+v.", err)
			}
		}

		// Write the result to chanel
		out <- c.rs
		log.Slogger.Infof("Exit goroutine.")
	}()
	pid, err := c.getPid()
	if err != nil {
		return
	}
	_, err = c.check(pid)
	if err != nil {
		return
	}
}

// Check the process status according to the pid number
// R: Running S: Sleep T: Stop I: Idle
// Z: Zombie W: Wait L: Lock
func (c *Check) check(pid int32) (string, error) {
	var err error
	defer func() {
		if err != nil {
			c.rs.AppendFailedStep(stepNameCheck, err)
		} else {
			c.rs.AppendSuccessStep(stepNameCheck)
		}
	}()
	p, err := process.NewProcess(pid)
	if err != nil {
		err = errors.WithStack(err)
		return "", err
	}

	stat, err := p.Status()
	if err != nil {
		err = errors.WithStack(err)
		return "", err
	}
	return stat, nil
}

// Read the pid file to get the pid
func (c *Check) getPid() (int32, error) {

	var err error

	// Open file
	pidFile, err := os.Open(c.PidFile)
	if err != nil {
		err = errors.WithStack(err)
		return 0, err
	}

	defer func() {
		if err != nil {
			c.rs.AppendFailedStep(stepNameGetPid, err)
		} else {
			c.rs.AppendSuccessStep(stepNameGetPid)
		}

		if err := pidFile.Close(); err != nil{
			log.Slogger.Errorf("*File Close Error: %s, File: %s", err.Error(), pidFile.Name())
		}
	}()

	if !afis.IsFile(c.PidFile) {
		err = errors.WithStack(NewPathError(c.PidFile, "Check PidFile Faild"))
		return 0, err
	}

	// Read
	content, err := ioutil.ReadAll(pidFile)
	if err != nil {
		err = errors.WithStack(err)
		return 0, err
	}

	pidInt, err := strconv.Atoi(strings.TrimSpace(string(content)))
	if err != nil {
		err = errors.WithStack(err)
		return 0, err
	}
	pid := int32(pidInt)
	return pid, nil
}
