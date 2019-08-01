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
	c, err := ftp.Dial(fu.client.Addr, ftp.DialWithTimeout(5*time.Second))

	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = fu.setPass()

	if err != nil{
		return nil, err
	}

	err = c.Login(fu.client.User, fu.client.Pass)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return c, nil
}

func (fu *FtpFileHandler) Upload() error {

	c, err := fu.conn()

	if err != nil {
		return errors.WithStack(err)
	}

	defer func() {
		_ = c.Quit()
		_ = c.Logout()
	}()

	// Do something with the FTP conn
	dirs := strings.Split(strings.Trim(fu.client.RelativePath, "/"), "/")

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

	file, err := os.Open(fu.client.Src)
	if err != nil {
		return errors.WithStack(err)
	}
	defer file.Close()
	err = c.Stor(filepath.Base(fu.client.Src), file)
	if err != nil {
		return errors.WithStack(err)
	}

	log.Slogger.Debugf("upload file sucess: %s", fu.client.Src)
	return nil
}

func (fu *FtpFileHandler) Get() (string, error){

	c, err := fu.conn()

	if err != nil {
		return "", errors.WithStack(err)
	}

	defer func() {
		_ = c.Quit()
		_ = c.Logout()
	}()

	begin := time.Now()
	//getcode
	tmpdir, err := ioutil.TempDir("", "rol_")
	if err != nil {
		return "", errors.WithStack(NewPathError("/tmp/dep_", err.Error()))
	}

	basedir := filepath.Dir(fu.client.RelativePath)

	downFile := filepath.Base(fu.client.RelativePath)

	err = c.ChangeDir(basedir)

	dir, _ := c.CurrentDir()

	log.Slogger.Debugf("current dir : %s", dir)

	if err != nil {
		return "", errors.WithStack(err)
	}

	resp, err := c.Retr(downFile)

	if err != nil {
		return "", errors.WithStack(err)
	}

	outFile, err := os.Create(filepath.Join(tmpdir, downFile))

	if err != nil {
		return "", errors.WithStack(err)
	}

	defer outFile.Close()

	_, err = io.Copy(outFile, resp)

	if err != nil {
		return "", errors.WithStack(err)
	}

	log.Slogger.Debugf("download file sucess: %s", filepath.Join(tmpdir, downFile))

	err = afis.Unzip(filepath.Join(tmpdir, downFile), tmpdir)

	if err != nil {
		return "", errors.WithStack(err)
	}

	log.Slogger.Debugf("unzip file sucess: %s", filepath.Join(tmpdir, downFile))

	elapsed := time.Since(begin)

	log.Slogger.Infof("download elapsed: ", elapsed)

	return tmpdir, nil
}
