/*
@Time : 19-5-6 下午1:52
@Author : liupeng
@File : rss.go
*/

package rss

import (
	"github.com/glory-cd/agent/common"
	"github.com/glory-cd/agent/executor"
	"github.com/glory-cd/utils/log"
	"github.com/pkg/errors"
)

type Rss struct {
	executor.Driver
}

func NewRss (ed executor.Driver) *Rss {
	newr := new(Rss)
	newr.Driver = ed
	return newr
}

func (r *Rss) deferHandleFunc(err *error, rs *executor.Result) {
	// Assert that the interface type of err is CoulsonError
	if *err != nil {
		rs.ReturnCode = common.ReturnCodeFailed
		rs.ReturnMsg = (*err).Error()
		if ce, ok := errors.Cause(*err).(executor.CoulsonError); ok {
			log.Slogger.Errorf("encounter an error:%+v, the kv is: %s", *err, ce.Kv())
		} else {
			log.Slogger.Errorf("encounter an error:%+v.", *err)
		}
	}
	log.Slogger.Infof("Exit goroutine.")
}

func (r *Rss) Exec(rs *executor.Result) {

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
	defer r.deferHandleFunc(&err, rs)
	rs.Identiy = r.Identiy

	switch r.OP {
	case common.OperateSTA:
		err = r.start(rs)
	case common.OperateSHU:
		err = r.shutdown(rs)
	case common.OperateRES:
		err = r.restart(rs)
	}
}

// Start the program
func (r *Rss) start(rs *executor.Result) error {
	var err error
	defer func() {
		if err != nil {
			rs.AppendFailedStep(executor.StepStart, err)
		} else {
			rs.AppendSuccessStep(executor.StepStart)
		}
	}()
	// Run the startup command
	err = r.RunCMD(r.StartCMD)
	if err != nil {
		return err
	}
	// Achieve the path of register script
	cmdMetaScript, err := r.GetMetaScript()
	if err != nil {
		return err
	}
	// Run the registration script
	err = r.RunCMD(cmdMetaScript)
	if err != nil {
		return err
	}

	return nil
}

// Shutdown the program
func (r *Rss) shutdown(rs *executor.Result) error {
	var err error
	defer func() {
		if err != nil {
			rs.AppendFailedStep(executor.StepStop, err)
		} else {
			rs.AppendSuccessStep(executor.StepStop)
		}
	}()

	err = r.RunCMD(r.StopCMD)
	if err != nil {
		return err
	}
	return nil
}

// Restart the program
func (r *Rss) restart(rs *executor.Result) error {
	err := r.shutdown(rs)
	if err != nil {
		return err
	}
	err = r.start(rs)
	if err != nil {
		return err
	}

	return nil
}
