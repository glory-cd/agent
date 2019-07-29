package executor

type Uploader interface {
	Upload() error

	// SetClient allows a getter to know it's client
	// in order to access client's Get functions or
	// progress tracking.
	SetClient(*Client)
}

func Upload(src, addr, utype, user, pass, path string) error {
	return (&Client{
		Src:          src,
		Addr:         addr,
		Type:         utype,
		User:         user,
		Pass:         pass,
		RelativePath: path,
	}).Upload()
}
