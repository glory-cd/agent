package executor

import (
	"github.com/glory-cd/utils/afis"
	"github.com/glory-cd/utils/log"
	"github.com/jlaffaye/ftp"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FtpFileHandler struct {
	baseHandler
}

//初始化连接和登录
func (fu *FtpFileHandler) conn() (*ftp.ServerConn, error){
	//连接FTP服务器
	c, err := ftp.Dial(fu.client.Addr, ftp.DialWithTimeout(5*time.Second))

	if err != nil {
		return nil, errors.WithStack(err)
	}
	err = fu.setPass() //解析密码
	if err != nil{
		return nil, err
	}
	//登录
	err = c.Login(fu.client.User, fu.client.Pass)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return c, nil
}

//上传
func (fu *FtpFileHandler) Upload() error {
	//初始化连接
	c, err := fu.conn()

	if err != nil {
		return errors.WithStack(err)
	}
	//延迟登出和关闭连接，如果发生错误记录日志
	defer func() {
		if err := c.Quit(); err != nil{
			log.Slogger.Errorf("FTP Quit error: %s", err.Error())
		}
		if err = c.Logout(); err != nil{
			log.Slogger.Errorf("FTP Login out error: %s", err.Error())
		}
	}()

	//分割路径
	dirs := strings.Split(strings.Trim(fu.client.RelativePath, "/"), "/")
	//为每层创建文件夹，并切换当前目录
	for _, v := range dirs {
		err = c.ChangeDir(v)
		if err != nil {
			_ = c.MakeDir(v)
			_ = c.ChangeDir(v)
		}
	}
	dir, err := c.CurrentDir()
	if err != nil {
		return errors.WithStack(err)
	}
	log.Slogger.Debugf("current dir : %s", dir)

	//打开文件准备上传
	file, err := os.Open(fu.client.Src)
	if err != nil {
		return errors.WithStack(err)
	}
	//文件延迟关闭
	defer func() {
		if err := file.Close(); err != nil{
			log.Slogger.Errorf("*File Close Error: %s, File: %s", err.Error(), file.Name())
		}
	}()
	//上传
	err = c.Stor(filepath.Base(fu.client.Src), file)
	if err != nil {
		return errors.WithStack(err)
	}

	log.Slogger.Debugf("upload file sucess: %s", fu.client.Src)
	return nil
}

//下载
func (fu *FtpFileHandler) Get() (string, error){
	//初始化连接
	c, err := fu.conn()

	if err != nil {
		return "", errors.WithStack(err)
	}
	//延迟登出和关闭连接，如果发生错误记录日志
	defer func() {
		if err := c.Quit(); err != nil{
			log.Slogger.Errorf("FTP Quit error: %s", err.Error())
		}
		if err = c.Logout(); err != nil{
			log.Slogger.Errorf("FTP Login out error: %s", err.Error())
		}
	}()

	begin := time.Now() //计时开始

	//创建临时存放文件夹
	tmpdir, err := ioutil.TempDir("", "get_")
	if err != nil {
		return "", errors.WithStack(NewPathError("/tmp/get_", err.Error()))
	}

	basedir := filepath.Dir(fu.client.RelativePath)

	downFile := filepath.Base(fu.client.RelativePath)
	//切换到文件所在目录
	err = c.ChangeDir(basedir)

	dir, _ := c.CurrentDir()

	log.Slogger.Debugf("current dir : %s", dir)

	if err != nil {
		return "", errors.WithStack(err)
	}
	//fetch the specified file from the remote FTP server, return ReadCloser
	resp, err := c.Retr(downFile)

	if err != nil {
		return "", errors.WithStack(err)
	}
	//创建目标文件，用于写入
	outFile, err := os.Create(filepath.Join(tmpdir, downFile))

	if err != nil {
		return "", errors.WithStack(err)
	}
	//延迟关闭文件描述符
	defer func() {
		if err := outFile.Close(); err != nil{
			log.Slogger.Errorf("*File Close Error: %s, File: %s", err.Error(), outFile.Name())
		}
	}()
	//copy 内容到目标文件
	_, err = io.Copy(outFile, resp)

	if err != nil {
		return "", errors.WithStack(err)
	}

	log.Slogger.Debugf("download file sucess: %s", filepath.Join(tmpdir, downFile))

	elapsed := time.Since(begin) //计时结束

	log.Slogger.Infof("download elapsed: ", elapsed)

	//解压下载的文件
	err = afis.Unzip(filepath.Join(tmpdir, downFile), tmpdir)

	if err != nil {
		return "", errors.WithStack(err)
	}

	log.Slogger.Debugf("unzip file sucess: %s", filepath.Join(tmpdir, downFile))


	return tmpdir, nil
}
