package executor

import (
	"fmt"
)

type Client struct {
	//source filepath
	Src string

	Addr string

	Type string

	Uploader Uploader

	User string

	Pass string

	RelativePath string
}

func (c *Client) Upload() error {

	switch c.Type {
	case "http":
		c.Uploader = new(HttpUploader)
	case "ftp":
		c.Uploader = new(FtpUploader)
	default:
		return fmt.Errorf("unsupported uploader: %s", c.Type)
	}

	c.Uploader.SetClient(c)

	err := c.Uploader.Upload()

	if err != nil {
		return err
	}

	return nil
}
