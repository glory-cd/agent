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

	User string

	Pass string

	RelativePath string
}

func (c *Client) init() error{
	switch c.Type {
	case "http":
		c.Handler = new(HttpFileHandler)
	case "ftp":
		c.Handler = new(FtpFileHandler)
	default:
		return fmt.Errorf("unsupported uploader: %s", c.Type)
	}

	c.Handler.SetClient(c)

	return nil
}

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