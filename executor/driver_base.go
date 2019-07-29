package executor

import (
	"github.com/auto-cdp/agent/common"
	"github.com/mholt/archiver"
	"github.com/pkg/errors"
)

type driver struct {
	*Task
	*Service
}


func (d *driver) backupService(filename, tmpdst, uploadpath string) error {
	// 压缩文件
	src := d.Dir

	err := archiver.Archive([]string{src}, tmpdst)
	if err != nil {
		return errors.WithStack(err)
	}
	//上传到文件服务器
	fileServer := common.Config().Upload
	err = Upload(
		tmpdst,
		fileServer.Addr,
		fileServer.Type,
		fileServer.UserName,
		fileServer.PassWord,
		uploadpath)

	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
