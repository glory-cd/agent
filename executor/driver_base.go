package executor

import (
	"fmt"
	"github.com/glory-cd/agent/common"
	"github.com/glory-cd/utils/afis"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type driver struct {
	*Task
	*Service
}

//备份dir到临时文件，并上传至文件服务器
func (d *driver) backupService(tmpdst, uploadpath string) error {
	//如果tmpdst目录不存在则创建
	if !afis.IsExists(filepath.Dir(tmpdst)) {
		err := os.MkdirAll(filepath.Dir(tmpdst), 0755)
		if err != nil {
			return err
		}
	}
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

	return strings.TrimSpace(string(path)), nil
}

func (d *driver) getBinPath(cmd string) (string, error) {
	var cmdpath string

	path := os.Getenv("PATH")

	pathSlice := strings.Split(path, string(os.PathListSeparator))

	for _, p := range pathSlice {
		fullcmd := filepath.Join(p, cmd)
		if afis.IsFile(fullcmd) && afis.IsExecutable(fullcmd) {
			cmdpath = fullcmd
			break
		}
	}

	if cmdpath == "" {
		return "", fmt.Errorf("command not found: %s", cmd)
	}
	return cmdpath, nil
}
