/*
@Time : 19-5-6 下午1:52
@Author : liupeng
@File : deploy.go
*/

package deploy

import (
	"bytes"
	"github.com/glory-cd/agent/common"
	"github.com/glory-cd/agent/executor"
	"github.com/glory-cd/utils/afis"
	"github.com/glory-cd/utils/log"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"sync"
	"time"
)

type deployError struct {
	Src        string `json:"src"`
	ServiceDir string `json:"servicedir"`
	errInf     string
}

func NewDeployError(thissrc, thisdir, thisErr string) *deployError {
	return &deployError{
		Src:        thissrc,
		ServiceDir: thisdir,
		errInf:     thisErr,
	}
}

func (de *deployError) Error() string {
	return de.errInf
}

func (de *deployError) Kv() string {
	return executor.Kv(de)
}

type Deploy struct {
	executor.Driver
	tempdir string
}

func NewDeploy(ed executor.Driver) *Deploy {
	newd := new(Deploy)
	newd.Driver = ed
	return newd
}

func (d *Deploy) Exec(rs *executor.Result) {
	log.Slogger.Infof("Begin to [Deploy] service：%s,%s", d.ServiceID, d.Dir)

	// Use the defer + closure to handle error returns and to clean up the temporary code
	var err error
	defer d.deferHandleFunc(&err, rs)

	rs.Identiy = d.Identiy

	err = d.checkenv(rs)
	if err != nil {
		return
	}

	// Initialize user directory, etc
	err = d.initEnv(rs)
	if err != nil {
		return
	}

	// Download the code
	err = d.downloadCode(rs)
	if err != nil {
		return
	}

	// Perform the deployment
	err = d.deploy(rs)
	if err != nil {
		return
	}
	// Register the service
	err = d.register(rs)
	if err != nil {
		return
	}
}

func (d *Deploy) deferHandleFunc(err *error, rs *executor.Result) {
	// Assert that the interface type of err is CoulsonError
	if *err != nil {
		rs.ReturnCode = common.ReturnCodeFailed
		rs.ReturnMsg = (*err).Error()
		if ce, ok := errors.Cause(*err).(executor.CoulsonError); ok {
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
	log.Slogger.Infof("Exit goroutine.")
}

// Deploy the service
func (d *Deploy) deploy(rs *executor.Result) error {
	var err error
	defer func() {
		if err != nil {
			rs.AppendFailedStep(executor.StepDeploy, err)
		} else {
			rs.AppendSuccessStep(executor.StepDeploy)
		}
	}()
	// Create the service directory
	err = os.Mkdir(d.Dir, 0755)
	if err != nil {
		return errors.WithStack(executor.NewPathError(d.Dir, err.Error()))
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
func (d *Deploy) checkenv(rs *executor.Result) error {

	// Check if the deployment path already exists and return an error if it does
	if afis.IsExists(d.Dir) {
		err := executor.NewPathError(d.Dir, "deploy path already exist")
		rs.AppendFailedStep(executor.StepCheckEnv, err)
		return errors.WithStack(err)
	}
	rs.AppendSuccessStep(executor.StepCheckEnv)
	return nil
}

// Stores a value indicating whether the user is being created
var syncChain sync.Map

// initEnv init environment, create user, create temporary directory etc.
func (d *Deploy) initEnv(rs *executor.Result) error {

	// Create a user if the user does not exist or wait for completion
	err := d.initUser(rs)

	if err != nil {
		return err
	}

	// Create a temporary code directory
	dir, err := ioutil.TempDir("", "dep_")
	if err != nil {
		rs.AppendFailedStep(executor.StepCreateTmpDir, err)
		return errors.WithStack(executor.NewPathError("/tmp/dep_", err.Error()))
	}
	rs.AppendSuccessStep(executor.StepCreateTmpDir)
	d.tempdir = dir
	log.Slogger.Infof("temp dir is : %s", d.tempdir)

	return nil
}

// initUser create user if user is not exist
// If a goroutine is doing creation, wait for it to complete
// return Timeout error if more than 10 seconds
func (d *Deploy) initUser(rs *executor.Result) error {

	if afis.IsUser(d.OsUser) {
		return nil
	}

	if _, ok := syncChain.LoadOrStore(d.OsUser, 1); !ok {
		//do task of creating user
		err := d.createUser(rs)

		if err != nil {
			return err
		}

		return nil

	}

	log.Slogger.Debugf("Begin to wait for user [%s] creation to complete.", d.OsUser)

	// check result every 10 ms
	begin := time.Now()
	for i := 0; i <= 1000; i++ {

		if afis.IsUser(d.OsUser) {

			log.Slogger.Debugf("Waiting for user [%s] creation took %s .", d.OsUser, time.Since(begin))
			return nil

		}

		time.Sleep(10 * time.Millisecond)
	}

	return errors.New("waiting for user creation timeout")
}

// createUser creates a user on the operating system
func (d *Deploy) createUser(rs *executor.Result) error {

	begin := time.Now()
	log.Slogger.Debugf("Begin to create user [%s].", d.OsUser)

	var err error
	defer func() {
		if err != nil {
			rs.AppendFailedStep(executor.StepCreateUser, err)
		} else {
			rs.AppendSuccessStep(executor.StepCreateUser)
		}

		elapsed := time.Since(begin)

		log.Slogger.Debugf("create user elapsed : %s", elapsed)
	}()

	// return if user has been created
	if afis.IsUser(d.OsUser) {
		return nil
	}

	cmdText, err := d.GetBinPath("useradd")
	if err != nil {
		return errors.WithStack(err)
	}

	// Assembling useradd's parameters
	options := []string{"-m"}
	if d.UserPass != "" {
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
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	if err != nil {
		return errors.Wrap(err, errStr)
	}

	log.Slogger.Debugf("create user [%s] success， OUT: %s", d.OsUser, outStr)

	return nil
}

// downloadCode gets the code from giving url
func (d *Deploy) downloadCode(rs *executor.Result) error {
	codedir, err := d.GetCode()
	if err != nil {
		rs.AppendFailedStep(executor.StepGetCode, err)
		return err
	}
	d.tempdir = codedir
	rs.AppendSuccessStep(executor.StepGetCode)
	return nil
}

//register this service to etcd
func (d *Deploy) register(rs *executor.Result) error {
	var err error
	defer func() {
		if err != nil {
			rs.AppendFailedStep(executor.StepRegister, err)
		} else {
			rs.AppendSuccessStep(executor.StepRegister)
		}
	}()
	// Achieve the path of register script
	cmdMetaScript, err := d.GetMetaScript()
	if err != nil {
		return err
	}
	// Run the registration script
	err = d.RunCMD(cmdMetaScript)
	if err != nil {
		return err
	}

	return nil
}
