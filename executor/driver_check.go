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
	"time"
)

type Check struct {
	driver
	rs Result
}

//构造ResultService
func (c *Check) constructRS(rcode common.ExecuteReturnCode, errstr string) {
	c.rs.AppendResultStep(
		1,
		"CHECK",
		rcode,
		errstr,
		time.Now().UnixNano(),
	)
}

func (c *Check) Exec(out chan<- Result) {
	log.Slogger.Infof("开始[CHECK]服务：%s,%s", c.ServiceID, c.Dir)
	var err error
	defer func() {

		//断言err的接口类型为CoulsonError
		if err != nil {
			c.rs.ReturnCode = common.ReturnCode_FAILED
			c.rs.ReturnMsg = err.Error()
			log.Slogger.Debugf("Result:%+v", c.rs)

			if ce, ok := errors.Cause(err).(CoulsonError); ok {
				log.Slogger.Errorf("encounter an error:%+v, the kv is: %s", err, ce.Kv())
			} else {
				log.Slogger.Errorf("encounter an error:%+v.", err)
			}
		}

		//结果写入chanel
		out <- c.rs
		log.Slogger.Infof("退出goroutine.")
	}()
	pid, err := c.getPid()
	if err != nil {
		c.constructRS(common.ReturnCode_FAILED, err.Error())
		return
	}
	status, err := c.check(pid)
	if err != nil {
		c.constructRS(common.ReturnCode_FAILED, err.Error())
		return
	}
	c.constructRS(common.ReturnCode_SUCCESS, status)
}

//根据pid号检查进程状态
// R: Running S: Sleep T: Stop I: Idle
// Z: Zombie W: Wait L: Lock
func (c *Check) check(pid int32) (string, error) {
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

//读取pid文件,获得pid
func (c *Check) getPid() (pid int32, e error) {
	if !afis.IsFile(c.PidFile) {
		e = errors.WithStack(NewPathError(c.PidFile, "Check PidFile Faild"))
		return
	}
	//打开文件
	pidFile, e := os.Open(c.PidFile)
	if e != nil {
		e = errors.WithStack(e)
		return
	}
	defer pidFile.Close()
	//读取
	content, e := ioutil.ReadAll(pidFile)
	if e != nil {
		e = errors.WithStack(e)
		return
	}

	pidInt, e := strconv.Atoi(strings.TrimSpace(string(content)))
	if e != nil {
		e = errors.WithStack(e)
		return
	}
	pid = int32(pidInt)
	return pid, nil
}
