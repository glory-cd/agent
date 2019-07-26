package afis

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/hashicorp/go-getter"
	"github.com/satori/go.uuid"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"syscall"
)

func GetMd5String(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func WriteUUID2File(uuidfile string) error {
	fpath := filepath.Dir(uuidfile)
	if ! IsDir(fpath){
		err := os.MkdirAll(fpath, 0755)
		if err != nil {
			return err
		}
	}
	u1 := uuid.Must(uuid.NewV4())
	err := ioutil.WriteFile(uuidfile, []byte(u1.String()), 0600)
	if err != nil {
		return err
	}
	return nil
}

func ReadUUIDFromFile(uuidfile string) (string, error) {
	uuidbyte, err := ioutil.ReadFile(uuidfile)
	if err != nil {
		return "", err
	}
	return string(uuidbyte), nil
}

func GetLocalIP() ([]string, error) {
	var iplist []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				iplist = append(iplist, ipnet.IP.String())
			}
		}
	}
	return iplist, nil
}

func GetHostName() (string, error) {
	hn, err := os.Hostname()
	if err != nil {
		return "", err
	}
	return hn, nil
}

//从http链接下载代码
//http://lp:0124@127.0.0.1/hfp/tst.tar.gz
func DownloadCode(dst, src string) error {
	err := getter.GetAny(dst, src)
	if err != nil {
		return err
	}

	return nil
}

//复制单个文件
func CopyFile(src, dst string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}

//复制整个目录
func CopyDir(src, dst string) error {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	if fds, err = ioutil.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = CopyDir(srcfp, dstfp); err != nil {
				return err
			}
		} else {
			if err = CopyFile(srcfp, dstfp); err != nil {
				return err
			}
		}
	}
	return nil
}

//更改文件属主
func ChownFile(fname, uname string) error {
	suser, err := user.Lookup(uname)
	if err != nil {
		return err
	}
	uid, _ := strconv.Atoi(suser.Uid)
	gid, _ := strconv.Atoi(suser.Gid)
	err = os.Chown(fname, uid, gid)
	if err != nil {
		return err
	}
	return nil
}

//递归更改整个目录属主
func ChownDirR(dname, uname string) error {

	var err error
	var fds []os.FileInfo

	suser, err := user.Lookup(uname)
	if err != nil {
		return err
	}
	uid, _ := strconv.Atoi(suser.Uid)
	gid, _ := strconv.Atoi(suser.Gid)
	err = os.Chown(dname, uid, gid)

	if fds, err = ioutil.ReadDir(dname); err != nil {
		return err
	}

	for _, fd := range fds {
		srcfp := path.Join(dname, fd.Name())

		if fd.IsDir() {
			if err = ChownDirR(srcfp, uname); err != nil {
				return err
			}
		} else {
			if err = ChownFile(srcfp, uname); err != nil {
				return err
			}
		}
	}
	return nil
}

//递归更改整个目录权限
func ChmodDirR(dname string, mode os.FileMode) error {

	var err error
	var fds []os.FileInfo

	err = os.Chmod(dname, mode)

	if fds, err = ioutil.ReadDir(dname); err != nil {
		return err
	}

	for _, fd := range fds {
		srcfp := path.Join(dname, fd.Name())

		if fd.IsDir() {
			if err = ChmodDirR(srcfp, mode); err != nil {
				return err
			}
		} else {
			if err = os.Chmod(srcfp, mode); err != nil {
				return err
			}
		}
	}
	return nil
}

//检查文件的属主与给定的user是否匹配
func CheckFileOwner(file, uname string) bool {
	info, err := os.Stat(file)
	if err != nil {
		return false
	}
	fuid := info.Sys().(*syscall.Stat_t).Uid
	suser, err := user.Lookup(uname)
	if err != nil {
		return false
	}
	uid, _ := strconv.Atoi(suser.Uid)

	if fuid == uint32(uid) {
		return true
	}

	return false
}

//删除目录中的内容
func RemoveContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

// readDirNames reads the directory named by dirname and returns
// a sorted list of directory entries.
func readDirNames(dirname string) ([]string, error) {
	f, err := os.Open(dirname)
	if err != nil {
		return nil, err
	}
	names, err := f.Readdirnames(-1)
	f.Close()
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	return names, nil
}

func WalkOnce(root string, walkFn filepath.WalkFunc) error {
	info, err := os.Lstat(root)
	if info != nil {
		err = walkOnce(root, info, walkFn)
	}
	if err == filepath.SkipDir {
		return nil
	}
	return err
}

// 仅walk path, calling walkFn.
func walkOnce(path string, info os.FileInfo, walkFn filepath.WalkFunc) error {
	if !info.IsDir() {
		return walkFn(path, info, nil)
	}

	names, err := readDirNames(path)

	if err != nil {
		return err
	}

	for _, name := range names {
		filename := filepath.Join(path, name)
		fileInfo, err := os.Lstat(filename)
		if err != nil {
			if err := walkFn(filename, fileInfo, err); err != nil && err != filepath.SkipDir {
				return err
			}
		}

		err = walkFn(filename, fileInfo, nil)

		if err != nil {
			return err
		}
	}
	return nil
}

// 判断所给路径文件/文件夹是否存在
func IsExists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

// 判断所给路径是否为文件夹
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// 判断所给路径是否为文件
func IsFile(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !s.IsDir()
}

//判断用户是否存在
func IsUser(uname string) bool {
	_, err := user.Lookup(uname)
	if err != nil {
		return false
	}
	return true
}
