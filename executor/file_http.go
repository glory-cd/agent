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

	// Client is the http.Client to use for Get requests.
	// This defaults to a cleanhttp.DefaultClient if left unset.
	Client *http.Client
}

func (hu *HttpFileHandler) Upload() error {

	begin := time.Now()

	err := hu.setPass()

	if err != nil{
		return err
	}

	src := hu.client.Src
	file, err := os.Open(src)
	if err != nil {
		return errors.WithStack(err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(src))
	if err != nil {
		return errors.WithStack(err)
	}
	_, err = io.Copy(part, file)

	err = writer.Close()
	if err != nil {
		return errors.WithStack(err)
	}

	log.Slogger.Debugf("upload request url : %s \n", hu.newPostUrl())

	req, err := http.NewRequest("POST", hu.newPostUrl(), body)

	if err != nil {
		return errors.WithStack(err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	hu.Client = cleanhttp.DefaultClient()

	resp, err := hu.Client.Do(req)

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
	begin := time.Now()

	err := hu.setPass()

	if err != nil{
		return "", err
	}
	//getcode
	//创建临时存放代码目录
	tmpdir, err := ioutil.TempDir("", "dep_")
	if err != nil {
		return "", errors.WithStack(NewPathError("/tmp/dep_", err.Error()))
	}

	//从url获取代码
	err = afis.DownloadCode(tmpdir, hu.newPostUrl())
	if err != nil {
		return "", errors.WithStack(NewGetCodeError(hu.newPostUrl(), err.Error()))
	}

	elapsed := time.Since(begin)

	log.Slogger.Infof("download elapsed: ", elapsed)

	return tmpdir, nil
}

func (hu *HttpFileHandler) newPostUrl() string {
	requestURL := new(url.URL)

	requestURL.Scheme = "http"

	requestURL.User = url.UserPassword(hu.client.User, hu.client.Pass)

	requestURL.Host = hu.client.Addr

	requestURL.Path += hu.client.RelativePath

	return requestURL.String()
}
