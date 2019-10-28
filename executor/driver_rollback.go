package executor

import (
	"github.com/glory-cd/agent/common"
	"github.com/glory-cd/utils/afis"
	"github.com/glory-cd/utils/log"
	"github.com/pkg/errors"
	"os"
	"path"
	"path/filepath"
)

type Roll struct {
	driver
	tmpdir string
	rs     Result
}

func (r *Roll) Exec(out chan<- Result) {
	log.Slogger.Infof("Begin to [ROLLBACK] service：%s,%s", r.ServiceID, r.Dir)
	var err error
	defer func() {
		// Assert that the interface type of err is CoulsonError
		if err != nil {
			r.rs.ReturnCode = common.ReturnCodeFailed
			r.rs.ReturnMsg = err.Error()
			log.Slogger.Debugf("Result:%+v", r.rs)
			if ce, ok := errors.Cause(err).(CoulsonError); ok {
				log.Slogger.Errorf("encounter an error:%+v, the kv is: %s", err, ce.Kv())
			} else {
				log.Slogger.Errorf("encounter an error:%+v.", err)
			}
		}

		// Write the result to chanel
		out <- r.rs
		log.Slogger.Infof("Exit goroutine.")
	}()

	err = r.getCode()

	if err != nil {
		r.rs.AppendFailedStep(stepNameGetCode, err)
		return
	}
	r.rs.AppendSuccessStep(stepNameGetCode)

	err = r.rollBack()

	if err != nil {
		r.rs.AppendFailedStep(stepNameRoll, err)
		return
	}
	r.rs.AppendSuccessStep(stepNameRoll)
}

// Download a backup copy from the file server
func (r *Roll) getCode() error {
	// Get the backup path that on the file server
	relativePath, err := r.readServiceVerion()

	if err != nil {

		return err
	}

	// Get a code backup from the file server
	fileServer := common.Config().FileServer
	dir, err := Get(fileServer, relativePath)

	if err != nil {
		return err
	}

	r.tmpdir = dir

	return nil
}

func (r *Roll) rollBack() error {
	//Return error if the file you want to delete belongs to a different user
	if !afis.CheckFileOwner(r.Dir, r.OsUser) {
		return errors.WithStack(NewFileOwnerError(r.Dir, r.OsUser, "file and owner does not match"))
	}
	// Delete the contents of the service directory
	err := os.RemoveAll(r.Dir)
	if err != nil {

		return errors.WithStack(err)
	}
	// Build the path and copy only the contents of the code directory
	// not including the code directory itself
	src := path.Join(r.tmpdir, filepath.Base(r.Dir))
	err = afis.CopyDir(src, r.Dir)
	if err != nil {
		return errors.WithStack(NewDeployError(src, r.Dir, err.Error()))
	}
	// Change the owner
	err = afis.ChownDirR(r.Dir, r.OsUser)
	if err != nil {
		return err
	}

	return nil
}
