package executor

import (
	"fmt"
	"github.com/pkg/errors"
)

//Client is a fileserver client
type Client struct {
	//source filepath
	Src string

	Addr string

	Type string

	Handler FileHandler
	//the username for http basic authorization or FTP.
	// if s3 is used, username is AWSAccessKeyId.
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

// Initialize the handler and set up a client for the handler
func (c *Client) init() error {
	switch c.Type {
	case "http":
		c.Handler = new(HTTPFileHandler)
	case "ftp":
		c.Handler = new(FTPFileHandler)
	case "s3":
		c.Handler = new(S3FileHandler)
	default:
		return fmt.Errorf("unsupported uploader: %s", c.Type)
	}

	c.Handler.SetClient(c)

	return nil
}

// Upload does upload actions
func (c *Client) Upload() error {

	err := c.init()

	if err != nil {
		return errors.WithStack(err)
	}
	err = c.Handler.Upload()

	if err != nil {
		return err
	}

	return nil
}

// Get does download actions
func (c *Client) Get() (string, error) {

	err := c.init()

	if err != nil {
		return "", errors.WithStack(err)
	}
	dir, err := c.Handler.Get()

	if err != nil {
		return "", err
	}

	return dir, nil
}
