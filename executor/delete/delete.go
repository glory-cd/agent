package delete

import (
	"github.com/glory-cd/agent/common"
	"github.com/glory-cd/agent/executor"
	"github.com/glory-cd/utils/log"
	"github.com/pkg/errors"
)

// Delete type
type Delete struct {
	executor.Driver
}

// NewDelete Instantiates a Delete
func NewDelete(ed executor.Driver) *Delete {
	newr := new(Delete)
	newr.Driver = ed
	return newr
}

// Exec is the main function that performs a service delete
func (r *Delete) Exec(rs *executor.Result) {
	log.Slogger.Infof("Begin to [Delete] service：%s,%s", r.ServiceID, r.Dir)
	var err error
	defer func() {
		// Assert that the interface type of err is CoulsonError
		if err != nil {
			rs.ReturnCode = common.ReturnCodeFailed
			rs.ReturnMsg = err.Error()
			log.Slogger.Debugf("Result:%+v", rs)
			if ce, ok := errors.Cause(err).(executor.CoulsonError); ok {
				log.Slogger.Errorf("encounter an error:%+v, the kv is: %s", err, ce.Kv())
			} else {
				log.Slogger.Errorf("encounter an error:%+v.", err)
			}
		}

		log.Slogger.Infof("Exit goroutine.")
	}()

	rs.Identiy = r.Identiy

	err = r.DeleteService()

	if err != nil {
		rs.AppendFailedStep(executor.StepDelete, err)
		return
	}
	rs.AppendSuccessStep(executor.StepDelete)
}
