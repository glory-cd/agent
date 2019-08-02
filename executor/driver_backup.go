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

//构造ResultService
func (b *Backup) constructRS(rcode common.ExecuteReturnCode, errstr string) {
	b.rs.AppendResultStep(
		1,
		"BACKUP",
		rcode,
		errstr,
		time.Now().UnixNano(),
	)
}

func (b *Backup) Exec(out chan<- Result) {
	log.Slogger.Infof("开始[BACKUP]服务：%s,%s", b.ServiceID, b.Dir)
	var err error
	defer func() {
		//断言err的接口类型为CoulsonError
		if err != nil {
			b.rs.ReturnCode = common.ReturnCode_FAILED
			b.rs.ReturnMsg = err.Error()
			log.Slogger.Debugf("Result:%+v", b.rs)
			if ce, ok := errors.Cause(err).(CoulsonError); ok {
				log.Slogger.Errorf("encounter an error:%+v, the kv is: %s", err, ce.Kv())
			} else {
				log.Slogger.Errorf("encounter an error:%+v.", err)
			}
		}

		//结果写入chanel
		out <- b.rs
		log.Slogger.Infof("退出goroutine.")
	}()

	//构建临时目标文件和上传路径
	filename := filepath.Base(b.Dir) + time.Now().Format("20060102150405.00000") + ".zip"
	dst := filepath.Join("/tmp/backup", filename)
	upath := filepath.Join(common.AgentID, b.ServiceID)
	//备份并上传
	err = b.backupService(dst, upath)

	if err != nil{
		b.constructRS(common.ReturnCode_FAILED, err.Error())
		return
	}

	b.constructRS(common.ReturnCode_SUCCESS, common.ReturnOKMsg)
}
