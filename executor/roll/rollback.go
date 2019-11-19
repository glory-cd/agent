package roll

import (
	"github.com/glory-cd/agent/common"
	"github.com/glory-cd/agent/executor"
	"github.com/glory-cd/utils/afis"
	"github.com/glory-cd/utils/log"
	"github.com/pkg/errors"
	"os"
	"path"
	"path/filepath"
)

type Roll struct {
	executor.Driver
	tmpdir string
}

func NewRoll (ed executor.Driver) *Roll {
	newr := new(Roll)
	newr.Driver = ed
	return newr
}

func (r *Roll) Exec(rs *executor.Result) {
	log.Slogger.Infof("Begin to [ROLLBACK] service：%s,%s", r.ServiceID, r.Dir)
	var err error
	defer func() {
		// Assert that the interface type of err is CoulsonError
		if err != nil {
			rs.ReturnCode = common.ReturnCodeFailed
			rs.ReturnMsg = err.Error()
			log.Slogger.Debugf("Result:%+v", rs)
			if ce, ok := errors.Cause(err).(executor.CoulsonError); ok {
				log.Slogger.Errorf("encounter an error:%+v, the kv is: %s", err, ce.Kv())
			} else {
				log.Slogger.Errorf("encounter an error:%+v.", err)
			}
		}

		log.Slogger.Infof("Exit goroutine.")
	}()

	rs.Identiy = r.Identiy
	err = r.getCode()

	if err != nil {
		rs.AppendFailedStep(executor.StepGetCode, err)
		return
	}
	rs.AppendSuccessStep(executor.StepGetCode)

	err = r.rollBack()

	if err != nil {
		rs.AppendFailedStep(executor.StepRoll, err)
		return
	}
	rs.AppendSuccessStep(executor.StepRoll)
}

// Download a backup copy from the file server
func (r *Roll) getCode() error {
	// Get the backup path that on the file server
	relativePath, err := r.ReadServiceVerion()

	if err != nil {

		return err
	}

	// Get a code backup from the file server
	fileServer := common.Config().FileServer
	dir, err := executor.Get(fileServer, relativePath)

	if err != nil {
		return err
	}

	r.tmpdir = dir

	return nil
}

func (r *Roll) rollBack() error {
	//Return error if the file you want to delete belongs to a different user
	if !afis.CheckFileOwner(r.Dir, r.OsUser) {
		return errors.WithStack(executor.NewFileOwnerError(r.Dir, r.OsUser, "file and owner does not match"))
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
		return errors.WithStack(executor.NewPathError(src, err.Error()))
	}
	// Change the owner
	err = afis.ChownDirR(r.Dir, r.OsUser)
	if err != nil {
		return err
	}

	return nil
}
