package executor

import (
	"bytes"
	"fmt"
	"github.com/glory-cd/utils/afis"
	"github.com/glory-cd/utils/log"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

type HttpFileHandler struct {
	baseHandler

	// HTTPClient is the http.HTTPClient to use for Get requests.
	// This defaults to a cleanhttp.DefaultClient if left unset.
	HTTPClient *http.Client
}

func (hu *HttpFileHandler) Upload() error {

	begin := time.Now() //计时开始
	//设置密码
	err := hu.setPass()
	if err != nil{
		return err
	}
	//打开文件
	src := hu.client.Src
	file, err := os.Open(src)
	if err != nil {
		return errors.WithStack(err)
	}
	//文件延迟关闭
	defer func() {
		if err := file.Close(); err != nil{
			log.Slogger.Errorf("*File Close Error: %s, File: %s", err.Error(), file.Name())
		}
	}()
	//创建multipart.Writer
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	//往Writer中写入文件类型的Part头信息
	part, err := writer.CreateFormFile("file", filepath.Base(src))
	if err != nil {
		return errors.WithStack(err)
	}
	//往Writer中写入文件类型的Part的Body信息
	_, err = io.Copy(part, file)
	//写入结尾符 --boudary--
	err = writer.Close()
	if err != nil {
		return errors.WithStack(err)
	}

	log.Slogger.Debugf("upload request url : %s \n", hu.newPostUrl())
	//创建http请求
	req, err := http.NewRequest("POST", hu.newPostUrl(), body)

	if err != nil {
		return errors.WithStack(err)
	}
	//设置header， Content-Type为multipart/form-data
	req.Header.Set("Content-Type", writer.FormDataContentType())

	//实例化http client
	hu.HTTPClient = cleanhttp.DefaultClient()
	//发起请求
	resp, err := hu.HTTPClient.Do(req)

	if err != nil {
		return errors.WithStack(err)
	}

	log.Slogger.Debugf("upload resp: %+v", resp)

	if resp.StatusCode != http.StatusOK {

		bodyBytes, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			return errors.WithStack(err)
		}

		return errors.WithStack(fmt.Errorf("upload statuscode: %d, mesage: %s", resp.StatusCode, string(bodyBytes)))

	}

	elapsed := time.Since(begin)

	log.Slogger.Infof("Upload elapsed: ", elapsed)

	return nil
}

func (hu *HttpFileHandler) Get() (string, error){

	begin := time.Now() //计时开始
	//设置密码
	err := hu.setPass()
	if err != nil{
		return "", err
	}
	//创建临时存放代码目录
	tmpdir, err := ioutil.TempDir("", "http_")
	if err != nil {
		return "", errors.WithStack(NewPathError("/tmp/http_", err.Error()))
	}
	//从url获取代码
	log.Slogger.Debugf("download from %s", hu.newPostUrl())
	err = afis.DownloadCode(tmpdir, hu.newPostUrl())
	if err != nil {
		return "", errors.WithStack(NewGetCodeError(hu.newPostUrl(), err.Error()))
	}

	elapsed := time.Since(begin) //计时结束

	log.Slogger.Infof("download elapsed: ", elapsed)

	return tmpdir, nil
}

//创建url.URL
func (hu *HttpFileHandler) newPostUrl() string {
	requestURL := new(url.URL)

	requestURL.Scheme = "http"

	requestURL.User = url.UserPassword(hu.client.User, hu.client.Pass)

	requestURL.Host = hu.client.Addr

	requestURL.Path += hu.client.RelativePath

	return requestURL.String()
}
