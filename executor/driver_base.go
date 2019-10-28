package executor

import (
	"fmt"
	"github.com/glory-cd/agent/common"
	"github.com/glory-cd/utils/afis"
	"github.com/glory-cd/utils/log"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type driver struct {
	*Task
	*Service
}

// Backup service dir to temporary file and upload to file server
func (d *driver) backupService(tmpdst, uploadpath string) error {
	// Create if the tmpdst directory does not exist
	if !afis.IsExists(filepath.Dir(tmpdst)) {
		err := os.MkdirAll(filepath.Dir(tmpdst), 0755)
		if err != nil {
			return err
		}
	}
	// Compressed files
	src := d.Dir
	err := afis.Zipit(src, tmpdst, "*.log")
	if err != nil {
		return errors.WithStack(err)
	}
	// Upload to the file server
	fileServer := common.Config().FileServer
	err = Upload(
		fileServer,
		tmpdst,
		uploadpath)

	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Read the PathFile file and get the backup path in the FileServer
func (d *driver) readServiceVerion() (string, error) {
	versionFile := filepath.Join(d.Dir, common.PathFile)
	path, err := ioutil.ReadFile(versionFile)
	if err != nil {
		return "", errors.WithStack(err)
	}

	return strings.TrimSpace(string(path)), nil
}

// Get the execution path of command
func (d *driver) getBinPath(cmd string) (string, error) {
	var cmdpath string

	path := os.Getenv("PATH")

	pathSlice := strings.Split(path, string(os.PathListSeparator))

	for _, p := range pathSlice {
		fullcmd := filepath.Join(p, cmd)
		if afis.IsFile(fullcmd) && afis.IsExecutable(fullcmd) {
			cmdpath = fullcmd
			break
		}
	}

	if cmdpath == "" {
		return "", fmt.Errorf("command not found: %s", cmd)
	}
	return cmdpath, nil
}

// Download code from FileServer
func (d *driver) getCode() (string, error) {
	//download code from url
	fileServer := common.Config().FileServer
	dir, err := Get(fileServer, d.RemoteCode)
	if err != nil {
		return "", err
	}

	log.Slogger.Infof("download code to %s", dir)
	return dir, nil
}
