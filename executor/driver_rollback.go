package executor

import (
	"github.com/glory-cd/agent/common"
	"github.com/glory-cd/utils/afis"
	"github.com/glory-cd/utils/log"
	"github.com/pkg/errors"
	"os"
	"path"
	"path/filepath"
	"time"
)

type Roll struct {
	driver
	tmpdir string
	rs     Result
}

//构造ResultService
func (r *Roll) constructRS(step int, rcode common.ExecuteReturnCode, errstr string) {
	r.rs.AppendResultStep(
		step,
		rollBackStepName[step],
		rcode,
		errstr,
		time.Now().UnixNano(),
	)
}

func (r *Roll) Exec(out chan<- Result) {
	log.Slogger.Infof("开始[ROLLBACK]服务：%s,%s", r.ServiceID, r.Dir)
	var err error
	defer func() {
		//断言err的接口类型为CoulsonError
		if err != nil {
			r.rs.ReturnCode = common.ReturnCode_FAILED
			r.rs.ReturnMsg = err.Error()
			log.Slogger.Debugf("Result:%+v", r.rs)
			if ce, ok := errors.Cause(err).(CoulsonError); ok {
				log.Slogger.Errorf("encounter an error:%+v, the kv is: %s", err, ce.Kv())
			} else {
				log.Slogger.Errorf("encounter an error:%+v.", err)
			}
		}

		//结果写入chanel
		out <- r.rs
		log.Slogger.Infof("退出goroutine.")
	}()

	err = r.getCode()

	if err != nil {
		return
	}

	err = r.rollBack()

	if err != nil {
		return
	}

}

//从文件服务器下载上一个备份副本
func (r *Roll) getCode() error {
	//获取文件服务器上的备份路径
	relativePath, err := r.readServiceVerion()

	if err != nil {
		r.constructRS(rollBackStepCodeGetCode, common.ReturnCode_FAILED, err.Error())
		return err
	}

	// 从文件服务器获取代码备份
	fileServer := common.Config().FileServer
	dir, err := Get(
		fileServer.Addr,
		fileServer.Type,
		fileServer.UserName,
		fileServer.PassWord,
		relativePath)

	if err != nil {
		r.constructRS(rollBackStepCodeGetCode, common.ReturnCode_FAILED, err.Error())
		return err
	}

	r.tmpdir = dir

	r.constructRS(rollBackStepCodeGetCode, common.ReturnCode_SUCCESS, common.ReturnOKMsg)

	return nil
}

func (r *Roll) rollBack() error {
	//如果要删除的文件属主与服务所在用户不同则直接返回*error
	if !afis.CheckFileOwner(r.Dir, r.OsUser) {
		r.constructRS(rollBackStepCodeRoll, common.ReturnCode_FAILED, "file and owner does not match")
		return errors.WithStack(
			NewFileOwnerError(r.Dir,
				r.OsUser,
				"file and owner does not match"))
	}
	//删除service目录下的内容
	err := os.RemoveAll(r.Dir)
	if err != nil {
		r.constructRS(rollBackStepCodeRoll, common.ReturnCode_FAILED, err.Error())
		return errors.WithStack(err)
	}
	//组装路径，仅复制代码目录中的内容，不包括代码目录本身
	src := path.Join(r.tmpdir, filepath.Base(r.Dir))
	err = afis.CopyDir(src, r.Dir)
	if err != nil {
		r.constructRS(rollBackStepCodeRoll, common.ReturnCode_FAILED, err.Error())
		return errors.WithStack(
			NewDeployError(
				src,
				r.Dir,
				err.Error(),
			),
		)
	}
	//更改属主
	err = afis.ChownDirR(r.Dir, r.OsUser)
	if err != nil {
		r.constructRS(rollBackStepCodeRoll, common.ReturnCode_FAILED, err.Error())
		return err
	}
	r.constructRS(rollBackStepCodeRoll, common.ReturnCode_SUCCESS, common.ReturnOKMsg)
	return nil
}
