package executor

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/glory-cd/utils/afis"
	"github.com/glory-cd/utils/log"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type S3FileHandler struct {
	baseHandler
	session *session.Session
}

func (sss *S3FileHandler) init() error {
	err := sss.setPass() //parsing password

	if err != nil {
		return errors.WithStack(err)
	}

	creds := credentials.NewStaticCredentials(sss.client.User, sss.client.Pass, "")

	config := &aws.Config{
		Region:           aws.String(sss.client.S3Region),
		S3ForcePathStyle: aws.Bool(true),
		Credentials:      creds,
	}
	// The session the S3 Downloader will use
	sss.session = session.Must(session.NewSession(config))

	return nil
}

func (sss *S3FileHandler) Upload() error {
	err := sss.init()
	if err != nil {
		return err
	}
	//upload
	begin := time.Now()
	// Create an uploader with the session and default options
	uploader := s3manager.NewUploader(sss.session)

	f, err := os.Open(sss.client.Src)
	if err != nil {
		return errors.WithStack(err)
	}

	key := filepath.Join(sss.client.RelativePath, filepath.Base(sss.client.Src))

	upParams := &s3manager.UploadInput{
		Bucket: aws.String(sss.client.S3Bucket),
		Key:    aws.String(key),
		Body:   f,
	}

	// Upload the file to S3.
	result, err := uploader.Upload(upParams, func(u *s3manager.Uploader) {
		u.PartSize = 10 * 1024 * 1024 // 10MB part size
		u.LeavePartsOnError = false   // Don't delete the parts if the upload fails.
		u.Concurrency = 3
	})

	if err != nil {
		return errors.WithStack(err)
	}
	elapsed := time.Since(begin)
	log.Slogger.Debugf("file uploaded to, %s\n elapsed: %s", aws.StringValue(&result.Location), elapsed)
	return nil
}

func (sss *S3FileHandler) Get() (string, error) {

	err := sss.init()
	if err != nil {
		return "", err
	}
	begin := time.Now() //timing begins

	// Create temporary storage folders
	tmpdir, err := ioutil.TempDir("", "s3_")
	if err != nil {
		return "", errors.WithStack(NewPathError("/tmp/s3_", err.Error()))
	}
	// Create a downloader with the session and default options
	downloader := s3manager.NewDownloader(sss.session)

	filename := filepath.Join(tmpdir, filepath.Base(sss.client.RelativePath))

	// Create a file to write the S3 Object contents to.
	f, err := os.Create(filename)
	if err != nil {
		return "", errors.WithStack(err)
	}

	// Write the contents of S3 Object to the file
	key := strings.TrimPrefix(sss.client.RelativePath, "/")
	n, err := downloader.Download(f, &s3.GetObjectInput{
		Bucket: aws.String(sss.client.S3Bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return "", errors.WithStack(err)
	}
	elapsed := time.Since(begin) //End of the timing

	log.Slogger.Infof("download elapsed: %s, %d bytes", elapsed, n)

	// Unzip the downloaded file
	err = afis.Unzip(filename, tmpdir)

	if err != nil {
		return "", errors.WithStack(err)
	}

	log.Slogger.Debugf("unzip file sucess: %s", filename)

	return tmpdir, nil
}
