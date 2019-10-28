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

// Initialize the connection and login
func (fu *FtpFileHandler) conn() (*ftp.ServerConn, error){
	// Connect to FTP server
	c, err := ftp.Dial(fu.client.Addr, ftp.DialWithTimeout(5*time.Second))

	if err != nil {
		return nil, errors.WithStack(err)
	}
	err = fu.setPass() //Parsing the password
	if err != nil{
		return nil, err
	}
	// Login
	err = c.Login(fu.client.User, fu.client.Pass)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return c, nil
}

// Upload
func (fu *FtpFileHandler) Upload() error {
	c, err := fu.conn()
	if err != nil {
		return errors.WithStack(err)
	}
	// Delay logout and close connections
	defer func() {
		if err := c.Quit(); err != nil{
			log.Slogger.Errorf("FTP Quit error: %s", err.Error())
		}
		if err = c.Logout(); err != nil{
			log.Slogger.Errorf("FTP Login out error: %s", err.Error())
		}
	}()

	// split path
	dirs := strings.Split(strings.Trim(fu.client.RelativePath, "/"), "/")
	// Create folders for each layer and switch the current directory
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

	// Open the file for upload
	file, err := os.Open(fu.client.Src)
	if err != nil {
		return errors.WithStack(err)
	}
	// delay close fd
	defer func() {
		if err := file.Close(); err != nil{
			log.Slogger.Errorf("*File Close Error: %s, File: %s", err.Error(), file.Name())
		}
	}()
	// upload
	err = c.Stor(filepath.Base(fu.client.Src), file)
	if err != nil {
		return errors.WithStack(err)
	}

	log.Slogger.Debugf("upload file sucess: %s", fu.client.Src)
	return nil
}

// Download
func (fu *FtpFileHandler) Get() (string, error){
	c, err := fu.conn()
	if err != nil {
		return "", errors.WithStack(err)
	}
	// Delay logout and close connections
	defer func() {
		if err := c.Quit(); err != nil{
			log.Slogger.Errorf("FTP Quit error: %s", err.Error())
		}
		if err = c.Logout(); err != nil{
			log.Slogger.Errorf("FTP Login out error: %s", err.Error())
		}
	}()

	begin := time.Now() //Timing begins

	// Create temporary storage folders
	tmpdir, err := ioutil.TempDir("", "get_")
	if err != nil {
		return "", errors.WithStack(NewPathError("/tmp/get_", err.Error()))
	}

	basedir := filepath.Dir(fu.client.RelativePath)

	downFile := filepath.Base(fu.client.RelativePath)
	// Switch to the directory where the file is
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
	// Creates a target file for writing
	outFile, err := os.Create(filepath.Join(tmpdir, downFile))

	if err != nil {
		return "", errors.WithStack(err)
	}
	// delay close fd
	defer func() {
		if err := outFile.Close(); err != nil{
			log.Slogger.Errorf("*File Close Error: %s, File: %s", err.Error(), outFile.Name())
		}
	}()
	// copy
	_, err = io.Copy(outFile, resp)

	if err != nil {
		return "", errors.WithStack(err)
	}

	log.Slogger.Debugf("download file sucess: %s", filepath.Join(tmpdir, downFile))

	elapsed := time.Since(begin) //End of the timing

	log.Slogger.Infof("download elapsed: ", elapsed)

	// Unzip the downloaded file
	err = afis.Unzip(filepath.Join(tmpdir, downFile), tmpdir)

	if err != nil {
		return "", errors.WithStack(err)
	}

	log.Slogger.Debugf("unzip file sucess: %s", filepath.Join(tmpdir, downFile))


	return tmpdir, nil
}
