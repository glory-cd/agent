package executor

import (
	"fmt"
	"github.com/pkg/errors"
)

type Client struct {
	//source filepath
	Src string

	Addr string

	Type string

	Handler FileHandler
	//the username for http basic authorization or FTP.
	// if s3 is used, Pass is AWSAccessKeyId.
	User string
	//the base64.StdEncoding.EncodeToString([]byte(password)) for http basic authorization or FTP.
	// if s3 is used, Pass is base64.StdEncoding.EncodeToString([]byte(AWSSecretAccessKey)).
	Pass string
	//RelativePath is path+file for Get
	//RelativePath is path for Upload
	RelativePath string
	// aws s3 region
	S3Region string
	// aws s3 bucket
	S3Bucket string
}

//初始化handler并为该handler设置client
func (c *Client) init() error{
	switch c.Type {
	case "http":
		c.Handler = new(HttpFileHandler)
	case "ftp":
		c.Handler = new(FtpFileHandler)
	case "s3":
		c.Handler = new(S3FileHandler)
	default:
		return fmt.Errorf("unsupported uploader: %s", c.Type)
	}

	c.Handler.SetClient(c)

	return nil
}

//上传
func (c *Client) Upload() error {

	err := c.init()

	if err != nil {
		return errors.WithStack(err)
	}
	//调用handler的upload函数
	err = c.Handler.Upload()

	if err != nil {
		return err
	}

	return nil
}

//下载
func (c *Client) Get() (string, error) {

	err := c.init()

	if err != nil {
		return "", errors.WithStack(err)
	}
	//调用handler的get函数
	dir, err := c.Handler.Get()

	if err != nil {
		return "", err
	}

	return dir, nil
}