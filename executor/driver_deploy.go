/*
@Time : 19-5-6 下午1:52
@Author : liupeng
@File : deploy.go
*/

package executor

import (
	"bytes"
	"github.com/glory-cd/agent/common"
	"github.com/glory-cd/utils/afis"
	"github.com/glory-cd/utils/log"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
)

type Deploy struct {
	driver
	rs      Result
	isuser  bool
	tempdir string
}

func (d *Deploy) Exec(out chan<- Result) {
	log.Slogger.Infof("Begin to [Deploy] service：%s,%s", d.ServiceID, d.Dir)

	// Use the defer + closure to handle error returns and to clean up the temporary code
	var err error
	defer d.deferHandleFunc(&err, out)

	err = d.checkenv()
	if err != nil {
		return
	}

	// Initialize user directory, etc
	err = d.initenv()
	if err != nil {
		return
	}

	// Download the code
	codedir, err := d.getCode()
	if err != nil {
		d.rs.AppendFailedStep(stepNameGetCode, err)
		return
	}
	d.tempdir = codedir
	d.rs.AppendSuccessStep(stepNameGetCode)

	// Perform the deployment
	err = d.deploy()
	if err != nil {
		d.rs.AppendFailedStep(stepNameDeploy, err)
		return
	}
	d.rs.AppendSuccessStep(stepNameDeploy)

}

func (d *Deploy) deferHandleFunc(err *error, out chan<- Result) {
	// Assert that the interface type of err is CoulsonError
	if *err != nil {
		d.rs.ReturnCode = common.ReturnCodeFailed
		d.rs.ReturnMsg = (*err).Error()
		if ce, ok := errors.Cause(*err).(CoulsonError); ok {
			log.Slogger.Errorf("encounter an error:%+v, the kv is: %s", *err, ce.Kv())
			// If deploy fails, delete the created directory
			if _, ok := ce.(*deployError); ok {
				err1 := os.RemoveAll(d.Dir)
				if err1 != nil {
					log.Slogger.Errorf("RemoveAll faild: %s.", err1.Error())
				}
			}
		} else {
			log.Slogger.Errorf("encounter an error:%+v.", *err)
		}
	}

	// Clean temporary directory
	if afis.IsExists(d.tempdir) {
		log.Slogger.Infof("clean temp dir %s.", d.tempdir)
		err2 := os.RemoveAll(d.tempdir)
		if err2 != nil {
			log.Slogger.Errorf("remove dir faild: %s.", err2.Error())
		}
	}
	// Write the result to chanel
	out <- d.rs
	log.Slogger.Infof("Exit goroutine.")
}

// Deploy the service
func (d *Deploy) deploy() error {
	// Create the service directory
	err := os.Mkdir(d.Dir, 0755)
	if err != nil {
		return errors.WithStack(NewPathError(d.Dir, err.Error()))
	}

	log.Slogger.Infof("create code dir: %s", d.Dir)

	// Build the path and copy only the contents of the code directory
	// not including the code directory itself
	src := path.Join(d.tempdir, d.ModuleName)
	err = afis.CopyDir(src, d.Dir)
	if err != nil {
		return errors.WithStack(NewDeployError(src, d.Dir, err.Error()))
	}
	log.Slogger.Infof("copy code from %s to %s successfully.", src, d.Dir)
	// Change the owner of the entire folder
	err = afis.ChownDirR(d.Dir, d.OsUser)
	if err != nil {
		return errors.WithStack(NewDeployError(src, d.Dir, err.Error()))
	}
	// Change permissions for the entire folder
	err = afis.ChmodDirR(d.Dir, 0755)
	if err != nil {
		return errors.WithStack(NewDeployError(src, d.Dir, err.Error()))
	}
	return nil
}

// Check the environment
func (d *Deploy) checkenv() error {
	// Check if the user exists
	if afis.IsUser(d.OsUser) {
		d.isuser = true
	}
	// Check if the deployment path already exists and return an error if it does
	if afis.IsExists(d.Dir) {
		err := NewPathError(d.Dir, "deploy path already exist")
		d.rs.AppendFailedStep(stepNameCheckEnv, err)
		return errors.WithStack(err)
	}
	d.rs.AppendSuccessStep(stepNameCheckEnv)
	return nil
}

func (d *Deploy) createUser() error{
	var err error
	defer func() {
		if err != nil {
			d.rs.AppendFailedStep(stepNameCreateUser, err)
		} else {
			d.rs.AppendSuccessStep(stepNameCreateUser)
		}
	}()

	cmdText, err := d.getBinPath("useradd")
	if err != nil {
		return errors.WithStack(err)
	}

	options := []string{"-m"}
	if d.UserPass != ""{
		withpass := []string{"-p", d.UserPass}
		options = append(options, withpass...)
	}
	options = append(options, d.OsUser)
	cmd := exec.Command(cmdText, options...)
	//Deal with stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	_, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	if err != nil {
		return errors.Wrap(err, errStr)
	}

	return nil
}

// Initialization environment
func (d *Deploy) initenv() error {
	// Create a user if the user does not exist
	if !d.isuser {
		err := d.createUser()
		if err != nil {
			return err
		}
		log.Slogger.Infof("create user %s success!", d.OsUser)
	} else {
		log.Slogger.Infof("The user %s already exists!", d.OsUser)
	}

	// Create a temporary code directory
	dir, err := ioutil.TempDir("", "dep_")
	if err != nil {
		d.rs.AppendFailedStep(stepNameCreateTmpDir, err)
		return errors.WithStack(NewPathError("/tmp/dep_", err.Error()))
	}
	d.rs.AppendSuccessStep(stepNameCreateTmpDir)
	d.tempdir = dir
	log.Slogger.Infof("temp dir is : %s", d.tempdir)

	return nil
}
