/*
@Time : 19-5-6 下午1:52
@Author : liupeng
@File : check.go
*/

package backup

import (
	"github.com/glory-cd/agent/common"
	"github.com/glory-cd/agent/executor"
	"github.com/glory-cd/utils/log"
	"github.com/pkg/errors"
	"path/filepath"
	"time"
)

type Backup struct {
	executor.Driver
}

func NewBackup (ed executor.Driver) *Backup {
	newr := new(Backup)
	newr.Driver = ed
	return newr
}

func (b *Backup) Exec(rs *executor.Result) {
	log.Slogger.Infof("Begin to [BACKUP] service：%s,%s", b.ServiceID, b.Dir)
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

	rs.Identiy = b.Identiy

	// Build temporary target files and upload paths
	filename := filepath.Base(b.Dir) + time.Now().Format("20060102150405.00000") + ".zip"
	dst := filepath.Join(common.TempBackupPath, filename)

	upath := filepath.Join(common.AgentID, b.ServiceID)
	// Backup and upload
	err = b.BackupService(dst, upath)

	if err != nil {
		rs.AppendFailedStep(executor.StepBackup, err)
		return
	}

	rs.AppendSuccessStep(executor.StepBackup)
}
