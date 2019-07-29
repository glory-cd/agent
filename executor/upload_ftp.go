package executor

import (
	"encoding/base64"
	"github.com/auto-cdp/utils/log"
	"github.com/jlaffaye/ftp"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FtpUploader struct {
	uploder
}

func (fu *FtpUploader) Upload() error {

	c, err := ftp.Dial(fu.client.Addr, ftp.DialWithTimeout(5*time.Second))

	if err != nil {
		return errors.WithStack(err)
	}

	defer c.Quit()

	decode, err := base64.StdEncoding.DecodeString(fu.client.Pass)
	if err != nil {
		return errors.WithStack(err)
	}

	fu.client.Pass = string(decode)

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
