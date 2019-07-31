package executor

import (
	"github.com/glory-cd/agent/common"
	"github.com/mholt/archiver"
	"github.com/pkg/errors"
	"io/ioutil"
	"path/filepath"
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