package executor

import (
	"encoding/base64"
	"github.com/pkg/errors"
)

type baseHandler struct {
	client *Client
}

func (u *baseHandler) SetClient(c *Client) { u.client = c }


func (u *baseHandler) setPass() error{
	//base64解密
	decode, err := base64.StdEncoding.DecodeString(u.client.Pass)
	if err != nil {
		return errors.WithStack(err)
	}

	u.client.Pass = string(decode)
	return nil
}