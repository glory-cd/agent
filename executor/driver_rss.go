/*
@Time : 19-5-6 下午1:52
@Author : liupeng
@File : rss.go
*/

package executor

import (
	"github.com/glory-cd/agent/common"
	"github.com/glory-cd/utils/log"
	"github.com/pkg/errors"
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
