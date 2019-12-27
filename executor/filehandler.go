package executor

import "github.com/glory-cd/agent/common"

// FileHandler is an interface that specifies how files are handled
type FileHandler interface {
	// upload file to file server
	Upload() error
	//download file from file server, it returns a dir where Stores the unzipped package
	Get() (string, error)
	// SetClient allows a filehandler to know it's client
	// in order to access client's Get&Upload functions or
	// progress tracking.
	SetClient(*Client)
}

// Upload builds a file client and calls Upload func
// src is file absolute path, path is a path without filename
func Upload(fs *common.StoreServer, src, path string) error {
	return (&Client{
		Src:          src,
		Addr:         fs.Addr,
		Type:         fs.Type,
		User:         fs.UserName,
		Pass:         fs.PassWord,
		S3Region:     fs.S3Region,
		S3Bucket:     fs.S3Bucket,
		RelativePath: path,
	}).Upload()
}

// Get builds a file client and calls Get func
// path is a path with filename
func Get(fs *common.StoreServer, path string) (string, error) {
	return (&Client{
		Addr:         fs.Addr,
		Type:         fs.Type,
		User:         fs.UserName,
		Pass:         fs.PassWord,
		S3Region:     fs.S3Region,
		S3Bucket:     fs.S3Bucket,
		RelativePath: path,
	}).Get()
}
