package executor

import (
	"github.com/glory-cd/agent/common"
	"github.com/glory-cd/utils/afis"
	"github.com/pkg/errors"
	"io/ioutil"
	"path/filepath"
)

type driver struct {
	*Task
	*Service
}

//备份dir到临时文件，并上传至文件服务器
func (d *driver) backupService(tmpdst, uploadpath string) error {
	// 压缩文件
	src := d.Dir
	err := afis.Zipit(src, tmpdst, "*.log")
	if err != nil {
		return errors.WithStack(err)
	}
	//上传到文件服务器
	fileServer := common.Config().FileServer
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

//读取PathFile文件，获取FileServer中的备份路径
func (d *driver) readServiceVerion() (string, error) {
	versionFile := filepath.Join(d.Dir, common.PathFile)
	path, err := ioutil.ReadFile(versionFile)
	if err != nil {
		return "", errors.WithStack(err)
	}

	return string(path), nil
}