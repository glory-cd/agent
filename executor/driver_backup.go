/*
@Time : 19-5-6 下午1:52
@Author : liupeng
@File : check.go
*/

package executor

import (
	"github.com/glory-cd/agent/common"
	"github.com/glory-cd/utils/log"
	"github.com/pkg/errors"
	"path/filepath"
	"time"
)

type Backup struct {
	driver
	rs Result
}

func (b *Backup) Exec(out chan<- Result) {
	log.Slogger.Infof("Begin to [BACKUP] service：%s,%s", b.ServiceID, b.Dir)
	var err error
	defer func() {
		// Assert that the interface type of err is CoulsonError
		if err != nil {
			b.rs.ReturnCode = common.ReturnCodeFailed
			b.rs.ReturnMsg = err.Error()
			log.Slogger.Debugf("Result:%+v", b.rs)
			if ce, ok := errors.Cause(err).(CoulsonError); ok {
				log.Slogger.Errorf("encounter an error:%+v, the kv is: %s", err, ce.Kv())
			} else {
				log.Slogger.Errorf("encounter an error:%+v.", err)
			}
		}

		// write the result to chanel
		out <- b.rs
		log.Slogger.Infof("Exit goroutine.")
	}()

	// Build temporary target files and upload paths
	filename := filepath.Base(b.Dir) + time.Now().Format("20060102150405.00000") + ".zip"
	dst := filepath.Join(common.TempBackupPath, filename)

	upath := filepath.Join(common.AgentID, b.ServiceID)
	// Backup and upload
	err = b.backupService(dst, upath)

	if err != nil {
		b.rs.AppendFailedStep(stepNameBackup, err)
		return
	}

	b.rs.AppendSuccessStep(stepNameBackup)
}
