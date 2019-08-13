package afis

import (
	"archive/zip"
	"crypto/md5"
	"encoding/hex"
	"github.com/glory-cd/utils/log"
	"github.com/hashicorp/go-getter"
	"github.com/pkg/errors"
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
	"strings"
	"syscall"
)

func GetMd5String(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func WriteUUID2File(uuidfile string) error {
	fpath := filepath.Dir(uuidfile)
	if !IsDir(fpath) {
		err := os.MkdirAll(fpath, 0755)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	u1 := uuid.Must(uuid.NewV4())
	err := ioutil.WriteFile(uuidfile, []byte(u1.String()), 0600)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func ReadUUIDFromFile(uuidfile string) (string, error) {
	uuidbyte, err := ioutil.ReadFile(uuidfile)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return string(uuidbyte), nil
}

func GetLocalIP() ([]string, error) {
	var iplist []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, errors.WithStack(err)
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
		return "", errors.WithStack(err)
	}
	return hn, nil
}

//从http链接下载代码
//http://lp:0124@127.0.0.1/hfp/tst.tar.gz
func DownloadCode(dst, src string) error {
	err := getter.GetAny(dst, src)
	if err != nil {
		return errors.WithStack(err)
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
		return errors.WithStack(err)
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil {
		return errors.WithStack(err)
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return errors.WithStack(err)
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return errors.WithStack(err)
	}
	return os.Chmod(dst, srcinfo.Mode())
}

//复制整个目录
func CopyDir(src, dst string) error {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(src); err != nil {
		return errors.WithStack(err)
	}

	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return errors.WithStack(err)
	}

	if fds, err = ioutil.ReadDir(src); err != nil {
		return errors.WithStack(err)
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = CopyDir(srcfp, dstfp); err != nil {
				return errors.WithStack(err)
			}
		} else {
			if err = CopyFile(srcfp, dstfp); err != nil {
				return errors.WithStack(err)
			}
		}
	}
	return nil
}

//更改文件属主
func ChownFile(fname, uname string) error {
	suser, err := user.Lookup(uname)
	if err != nil {
		return errors.WithStack(err)
	}
	uid, _ := strconv.Atoi(suser.Uid)
	gid, _ := strconv.Atoi(suser.Gid)
	err = os.Chown(fname, uid, gid)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

//递归更改整个目录属主
func ChownDirR(dname, uname string) error {

	var err error
	var fds []os.FileInfo

	suser, err := user.Lookup(uname)
	if err != nil {
		return errors.WithStack(err)
	}
	uid, _ := strconv.Atoi(suser.Uid)
	gid, _ := strconv.Atoi(suser.Gid)
	err = os.Chown(dname, uid, gid)

	if fds, err = ioutil.ReadDir(dname); err != nil {
		return errors.WithStack(err)
	}

	for _, fd := range fds {
		srcfp := path.Join(dname, fd.Name())

		if fd.IsDir() {
			if err = ChownDirR(srcfp, uname); err != nil {
				return errors.WithStack(err)
			}
		} else {
			if err = ChownFile(srcfp, uname); err != nil {
				return errors.WithStack(err)
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
		return errors.WithStack(err)
	}

	for _, fd := range fds {
		srcfp := path.Join(dname, fd.Name())

		if fd.IsDir() {
			if err = ChmodDirR(srcfp, mode); err != nil {
				return errors.WithStack(err)
			}
		} else {
			if err = os.Chmod(srcfp, mode); err != nil {
				return errors.WithStack(err)
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
		return errors.WithStack(err)
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return errors.WithStack(err)
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

// readDirNames reads the directory named by dirname and returns
// a sorted list of directory entries.
func readDirNames(dirname string) ([]string, error) {
	f, err := os.Open(dirname)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	names, err := f.Readdirnames(-1)
	f.Close()
	if err != nil {
		return nil, errors.WithStack(err)
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
	return errors.WithStack(err)
}

// 仅walk path, calling walkFn.
func walkOnce(path string, info os.FileInfo, walkFn filepath.WalkFunc) error {
	if !info.IsDir() {
		return walkFn(path, info, nil)
	}

	names, err := readDirNames(path)

	if err != nil {
		return errors.WithStack(err)
	}

	for _, name := range names {
		filename := filepath.Join(path, name)
		fileInfo, err := os.Lstat(filename)
		if err != nil {
			if err := walkFn(filename, fileInfo, err); err != nil && err != filepath.SkipDir {
				return errors.WithStack(err)
			}
		}

		err = walkFn(filename, fileInfo, nil)

		if err != nil {
			return errors.WithStack(err)
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

//解压zip
func Unzip(archive, target string) error {
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return errors.WithStack(err)
	}

	if err := os.MkdirAll(target, 0755); err != nil {
		return errors.WithStack(err)
	}

	for _, file := range reader.File {
		unzippath := filepath.Join(target, file.Name)
		if file.FileInfo().IsDir() {
			err := os.MkdirAll(unzippath, file.Mode())
			if err != nil {
				return errors.WithStack(err)
			}
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return errors.WithStack(err)
		}
		defer fileReader.Close()

		targetFile, err := os.OpenFile(unzippath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return errors.WithStack(err)
		}
		defer targetFile.Close()

		if _, err := io.Copy(targetFile, fileReader); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

//压缩为zip格式
func Zipit(source, target, filter string) error {
	var err error
	if isAbs := filepath.IsAbs(source); !isAbs {
		source, err = filepath.Abs(source) // 将传入路径直接转化为绝对路径
		if err != nil {
			return errors.WithStack(err)
		}
	}
	//创建zip包文件
	zipfile, err := os.Create(target)
	if err != nil {
		return errors.WithStack(err)
	}

	defer func() {
		if err := zipfile.Close(); err != nil {
			log.Slogger.Errorf("*File close error: %s, file: %s", err.Error(), zipfile.Name())
		}
	}()

	//创建zip.Writer
	zw := zip.NewWriter(zipfile)

	defer func() {
		if err := zw.Close(); err != nil {
			log.Slogger.Errorf("zipwriter close error: %s", err.Error())
		}
	}()

	info, err := os.Stat(source)
	if err != nil {
		return errors.WithStack(err)
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			return errors.WithStack(err)
		}

		//将遍历到的路径与pattern进行匹配
		ism, err := filepath.Match(filter, info.Name())

		if err != nil {
			return errors.WithStack(err)
		}
		//如果匹配就忽略
		if ism {
			return nil
		}
		//创建文件头
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return errors.WithStack(err)
		}

		if baseDir != "" {
			header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}
		//写入文件头信息
		writer, err := zw.CreateHeader(header)
		if err != nil {
			return errors.WithStack(err)
		}

		if info.IsDir() {
			return nil
		}
		//写入文件内容
		file, err := os.Open(path)
		if err != nil {
			return errors.WithStack(err)
		}

		defer func() {
			if err := file.Close(); err != nil {
				log.Slogger.Errorf("*File close error: %s, file: %s", err.Error(), file.Name())
			}
		}()
		_, err = io.Copy(writer, file)

		return errors.WithStack(err)
	})

	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

//check file is executable
func IsExecutable(path string) bool {
	//Filemode with execute permissions
	var exemode os.FileMode = 0111
	//get path filemode
	fileinfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	filemode := fileinfo.Mode()
	//& operation
	r := exemode & filemode
	if uint32(r) != 0 {
		return true
	}
	return false
}
