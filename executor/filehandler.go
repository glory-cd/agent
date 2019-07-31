package executor

type FileHandler interface {
	Upload() error

	Get() (string, error)

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


func Get(addr, utype, user, pass, path string) (string, error) {
	return (&Client{
		Addr:         addr,
		Type:         utype,
		User:         user,
		Pass:         pass,
		RelativePath: path,
	}).Get()
}