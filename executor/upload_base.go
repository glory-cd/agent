package executor

type uploder struct {
	client *Client
}

func (u *uploder) SetClient(c *Client) { u.client = c }
