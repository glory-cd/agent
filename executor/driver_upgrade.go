/*
@Time : 19-5-6 下午1:52
@Author : liupeng
@File : upgrade.go
*/

package executor

import (
	"github.com/glory-cd/agent/common"
	"github.com/glory-cd/utils/afis"
	"github.com/glory-cd/utils/log"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type Upgrade struct {
	driver
	rs             Result
	tmpdir         string
	backfile       string
	executepattern []string
}

func (u *Upgrade) Exec(out chan<- Result) {
	log.Slogger.Infof("开始升级服务：%s,%s", u.ServiceID, u.Dir)

	var err error
	defer u.deferHandleFunc(&err, out)

	//backup and upload
	err = u.backup()
	if err != nil {
		u.rs.AppendFailedStep(stepNameBackup, err)
		return
	}
	u.rs.AppendSuccessStep(stepNameBackup)

	//create temp dir to store code
	err = u.createTempDir()
	if err != nil {
		u.rs.AppendFailedStep(stepNameGetCode, err)
		return
	}
	//download code
	codedir, err := u.getCode()
	if err != nil {
		u.rs.AppendFailedStep(stepNameGetCode, err)
		return
	}
	u.tmpdir = codedir
	u.rs.AppendSuccessStep(stepNameGetCode)

	//verify code and service
	err = u.checkenv()
	if err != nil {
		u.rs.AppendFailedStep(stepNameCheckEnv, err)
		return
	}
	u.rs.AppendSuccessStep(stepNameCheckEnv)

	//perform an upgrade
	err = u.upgrade()
	if err != nil {
		u.rs.AppendFailedStep(stepNameUpgrade, err)
		return
	}
	u.rs.AppendSuccessStep(stepNameUpgrade)

}

func (u *Upgrade) deferHandleFunc(err *error, out chan<- Result) {
	if *err != nil {
		u.rs.ReturnCode = common.ReturnCodeFailed
		u.rs.ReturnMsg = (*err).Error()
		//assert the interface type(CoulsonError)
		if ce, ok := errors.Cause(*err).(CoulsonError); ok {
			log.Slogger.Errorf("encounter an error:%+v, the kv is: %s", *err, ce.Kv())

			//rollback if dealPatternDirs or dealPatternFiles fails
			if _, ok := ce.(*dealPatternError); ok {
				err1 := u.rollBack()
				if err1 != nil {
					log.Slogger.Errorf("rollBack faild: %s.", err1.Error())
				}
			}
		} else {
			log.Slogger.Errorf("encounter an error:%+v.", *err)
		}
	}

	//Clean temporary directory
	if afis.IsExists(u.tmpdir) {
		log.Slogger.Infof("clean temp dir %s.", u.tmpdir)
		err2 := os.RemoveAll(u.tmpdir)
		if err2 != nil {
			log.Slogger.Errorf("remove dir faild: %s.", err2.Error())
		}
	}
	//send result to chanel
	out <- u.rs
	log.Slogger.Infof("退出goroutine.")
}

//check && verify
func (u *Upgrade) checkenv() error {

	log.Slogger.Debugf("CustomPattern and CodePattern:%+v, %d, %+v, %d", u.CustomPattern, len(u.CustomPattern), u.CodePattern, len(u.CodePattern))

	if afis.ContainsString(u.CodePattern, "") || afis.ContainsString(u.CustomPattern, "") {
		log.Slogger.Errorf("Pattern has empty string.")
		return errors.New("pattern has empty string")
	}

	if u.CustomPattern != nil {
		u.executepattern = u.CustomPattern
	} else {
		u.executepattern = u.CodePattern
	}

	if u.executepattern == nil {
		return errors.New("pattern is nil")
	}
	//Check if the deployment path already exists
	if !afis.IsExists(u.Dir) {
		return errors.WithStack(NewPathError(u.Dir, "program is not exist"))
	}
	//Check if the corresponding path of CodePattern exists in the deployment directory
	for _, dcp := range u.executepattern {
		codepath := path.Join(u.Dir, dcp)
		if strings.ContainsAny(codepath, "*?[]") {
			codepath = filepath.Dir(codepath)
		}
		if !afis.IsExists(codepath) {
			return errors.WithStack(NewPathError(codepath, "codepath is not exist"))
		}

	}

	// Check if the corresponding path of CodePattern exists in the temporary code directory.
	// return *pathError if not
	for _, tcp := range u.executepattern {
		tcodepath := path.Join(u.tmpdir, u.ModuleName, tcp)
		if strings.ContainsAny(tcodepath, "*?[]") {
			tcodepath = filepath.Dir(tcodepath)
		}
		if !afis.IsExists(tcodepath) {
			return errors.WithStack(NewPathError(tcodepath, "patterndir is not exist"))
		}

	}
	log.Slogger.Info("checkenv  successfully.")

	return nil
}

func (u *Upgrade) createTempDir() error {
	dir, err := ioutil.TempDir("", "upg_")
	if err != nil {
		return errors.WithStack(NewPathError("/tmp/upg_", err.Error()))

	}
	u.tmpdir = dir
	log.Slogger.Infof("temp dir is : %s", u.tmpdir)
	return nil
}

func (u *Upgrade) upgrade() error {
	patterndirs, patternfiles, patterns := u.classifyPattern()
	log.Slogger.Debugw("The paths need to be precessed:", "patterndirs", patterndirs,
		"patternfiles", patternfiles, "patterns", patterns)

	//process dirs
	if patterndirs != nil {
		err := u.dealPatternDirs(patterndirs)
		if err != nil {
			return err
		}
	}
	//process files
	if patternfiles != nil {
		err := u.dealPatternFiles(patternfiles)
		if err != nil {
			return err
		}
	}
	//process patterns
	if patterns != nil {
		err := u.dealPatterns(patterns)
		if err != nil {
			return err
		}
	}

	//change the owner of the entire folder
	err := afis.ChownDirR(u.Dir, u.OsUser)
	if err != nil {
		return errors.WithStack(NewPathError(u.Dir, err.Error()))
	}

	//change permissions for the entire folder
	err = afis.ChmodDirR(u.Dir, 0755)
	if err != nil {
		return errors.WithStack(NewPathError(u.Dir, err.Error()))
	}

	return nil
}

//CodePattern was classified according to directories, files and Patterns
func (u *Upgrade) classifyPattern() (patterndirs, patternfiles, patterns []map[string]string) {
	var pdirs []map[string]string
	var pfiles []map[string]string
	var ppatterns []map[string]string

	log.Slogger.Debugf("ExecutePattern:%+v", u.executepattern)
	for _, cp := range u.executepattern {
		codepath := path.Join(u.Dir, cp)
		//Load both the pattern absolute path and the relative path into the map
		dm := make(map[string]string)
		dm["codepattern"] = cp
		dm["codepath"] = codepath
		if afis.IsDir(codepath) {
			pdirs = append(pdirs, dm)
			continue
		}

		if afis.IsFile(codepath) {
			pfiles = append(pfiles, dm)
			continue
		}
		//load the path that is neither a directory nor a file (real pattern) into the ppatterns
		ppatterns = append(ppatterns, dm)
	}
	return pdirs, pfiles, ppatterns
}

//Process directories in patterns
func (u *Upgrade) dealPatternDirs(pdirs []map[string]string) error {
	for _, pd := range pdirs {
		//Check if the owner matches
		if !afis.CheckFileOwner(pd["codepath"], u.OsUser) {
			return errors.WithStack(NewFileOwnerError(pd["codepath"], u.OsUser, "file and owner does not match"))
		}
		err := afis.RemoveContents(pd["codepath"])
		if err != nil {
			return errors.WithStack(NewdealPatternError(u.OsUser, "", pd["codepath"], err.Error()))
		}
		//copy direactory
		src := filepath.Join(u.tmpdir, u.ModuleName, pd["codepattern"])
		err = afis.CopyDir(src, pd["codepath"])
		if err != nil {
			return errors.WithStack(NewdealPatternError(u.OsUser, src, pd["codepath"], err.Error()))
		}
	}
	return nil
}

//Process files in patterns
func (u *Upgrade) dealPatternFiles(pfiles []map[string]string) error {
	for _, pf := range pfiles {
		//Check if the owner matches
		if !afis.CheckFileOwner(pf["codepath"], u.OsUser) {
			return errors.WithStack(NewFileOwnerError(pf["codepath"], u.OsUser, "file and owner does not match"))
		}
		err := os.Remove(pf["codepath"])
		if err != nil {
			return errors.WithStack(NewdealPatternError(u.OsUser, "", pf["codepath"], err.Error()))
		}
		//copy file
		src := filepath.Join(u.tmpdir, u.ModuleName, pf["codepattern"])
		err = afis.CopyFile(src, pf["codepath"])
		if err != nil {
			return errors.WithStack(NewdealPatternError(u.OsUser, src, pf["codepath"], err.Error()))
		}
	}
	return nil
}

//Process patterns
func (u *Upgrade) dealPatterns(ppatterns []map[string]string) error {
	var pdirs []map[string]string
	var pfiles []map[string]string
	for _, pp := range ppatterns {
		basedir := filepath.Dir(pp["codepath"])
		//just walk current pattern's root dir for once, and process according to file and directory
		err := afis.WalkOnce(basedir, func(path string, info os.FileInfo, err1 error) error {
			if err1 != nil {
				return errors.WithStack(NewdealPatternError(u.OsUser, "", basedir, err1.Error()))
			}
			//Check whether the path matches the pattern
			ism, err2 := filepath.Match(pp["codepath"], path)

			if err2 != nil {
				return errors.WithStack(NewdealPatternError(u.OsUser, pp["codepath"], path, err2.Error()))
			}
			//classify according to directories, files
			if ism {
				dm := make(map[string]string)
				dm["codepattern"] = filepath.Join(filepath.Dir(pp["codepattern"]), filepath.Base(path))
				dm["codepath"] = path

				if afis.IsDir(path) {
					pdirs = append(pdirs, dm)
				}

				if afis.IsFile(path) {
					pfiles = append(pfiles, dm)
				}
			}

			return nil
		})

		if err != nil {
			return err
		}
	}
	log.Slogger.Debugw("The paths have been matched:", "dirs", pdirs,
		"files", pfiles)
	//Process the matched directories
	err := u.dealPatternDirs(pdirs)
	if err != nil {
		return err
	}
	//Process the matched files
	err = u.dealPatternFiles(pfiles)
	if err != nil {
		return err
	}
	return nil
}

//backup and upload
func (u *Upgrade) backup() error {
	//build filename
	filename := filepath.Base(u.Dir) + time.Now().Format("20060102150405.00000") + ".zip"
	//build dst file path
	dst := filepath.Join(common.TempBackupPath, filename)
	//build upload path
	upath := filepath.Join(common.AgentID, u.ServiceID)
	//backup and upload
	err := u.backupService(dst, upath)

	if err != nil {
		return err
	}
	//build filepath with name that reside on file server
	remoteFilePath := filepath.Join(upath, filename)
	//local version file which contains remoteFilePath
	versionFile := filepath.Join(u.Dir, common.PathFile)
	err = ioutil.WriteFile(versionFile, []byte(remoteFilePath), 0644)

	if err != nil {
		return errors.WithStack(err)
	}
	//change owner
	err = afis.ChownFile(versionFile, u.OsUser)

	if err != nil {
		return errors.WithStack(err)
	}
	//record temporary backup file for rolling back
	u.backfile = dst
	return nil
}

//rollback
func (u *Upgrade) rollBack() error {
	//Check if the owner matches
	if !afis.CheckFileOwner(u.Dir, u.OsUser) {
		return errors.WithStack(NewFileOwnerError(u.Dir, u.OsUser, "file and owner does not match"))
	}
	err := os.RemoveAll(u.Dir)
	if err != nil {
		return errors.WithStack(err)
	}
	//unzip temporary backup file to basedir
	basedir := filepath.Dir(u.Dir)
	err = afis.Unzip(u.backfile, basedir)
	if err != nil {
		return err
	}
	err = afis.ChownDirR(u.Dir, u.OsUser)
	if err != nil {
		return err
	}
	return nil
}
