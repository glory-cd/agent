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

	begin := time.Now() //Timing begins
	// set password
	err := hu.setPass()
	if err != nil{
		return err
	}
	// open file
	src := hu.client.Src
	file, err := os.Open(src)
	if err != nil {
		return errors.WithStack(err)
	}
	// delay tp close fd
	defer func() {
		if err := file.Close(); err != nil{
			log.Slogger.Errorf("*File Close Error: %s, File: %s", err.Error(), file.Name())
		}
	}()
	// create multipart.Writer
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	// Writes Part header information for the file type to Writer
	part, err := writer.CreateFormFile("file", filepath.Base(src))
	if err != nil {
		return errors.WithStack(err)
	}
	// Writes Part Body information for the file type to Writer
	_, err = io.Copy(part, file)
	// Write end character --boudary--
	err = writer.Close()
	if err != nil {
		return errors.WithStack(err)
	}

	log.Slogger.Debugf("upload request url : %s \n", hu.newPostUrl())
	// Create an HTTP request
	req, err := http.NewRequest("POST", hu.newPostUrl(), body)

	if err != nil {
		return errors.WithStack(err)
	}
	// Set the content-type of header to multipart/form-data
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Instantiate the HTTP client
	hu.HTTPClient = cleanhttp.DefaultClient()
	// Make an HTTP request
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

	elapsed := time.Since(begin)//End of the timing

	log.Slogger.Infof("Upload elapsed: ", elapsed)

	return nil
}

func (hu *HttpFileHandler) Get() (string, error){

	begin := time.Now() //Timing begins
	//set password
	err := hu.setPass()
	if err != nil{
		return "", err
	}
	// Create temporary storage folders
	tmpdir, err := ioutil.TempDir("", "http_")
	if err != nil {
		return "", errors.WithStack(NewPathError("/tmp/http_", err.Error()))
	}
	// Get the code from the url
	log.Slogger.Debugf("download from %s", hu.newPostUrl())
	err = afis.DownloadCode(tmpdir, hu.newPostUrl())
	if err != nil {
		return "", errors.WithStack(NewGetCodeError(hu.newPostUrl(), err.Error()))
	}

	elapsed := time.Since(begin) //End of the timing

	log.Slogger.Infof("download elapsed: ", elapsed)

	return tmpdir, nil
}

//Build url.URL
func (hu *HttpFileHandler) newPostUrl() string {
	requestURL := new(url.URL)

	requestURL.Scheme = "http"

	requestURL.User = url.UserPassword(hu.client.User, hu.client.Pass)

	requestURL.Host = hu.client.Addr

	requestURL.Path += hu.client.RelativePath

	return requestURL.String()
}
