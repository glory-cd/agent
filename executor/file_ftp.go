package executor

import (
	"github.com/glory-cd/utils/log"
	"github.com/jlaffaye/ftp"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FtpFileHandler struct {
	baseHandler
}

func (fu *FtpFileHandler) Upload() error {

	c, err := ftp.Dial(fu.client.Addr, ftp.DialWithTimeout(5*time.Second))

	if err != nil {
		return errors.WithStack(err)
	}

	defer c.Quit()

	err = fu.setPass()

	if err != nil{
		return err
	}

	err = c.Login(fu.client.User, fu.client.Pass)
	if err != nil {
		return errors.WithStack(err)
	}

	defer c.Logout()

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

	return nil
}

func (fu *FtpFileHandler) Get() (string, error){
	return "", nil
}
