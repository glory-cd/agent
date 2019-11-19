/*
@Time : 19-5-6 下午1:52
@Author : liupeng
@File : upgrade.go
*/

package upgrade

import (
	"github.com/glory-cd/agent/common"
	"github.com/glory-cd/agent/executor"
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

type dealPatternError struct {
	Owner  string `json:"owner"`
	Src    string `json:"src"`
	Path   string `json:"path"`
	errInf string
}

// newDealPatternError return a *dealPatternError
func newDealPatternError(thisOwner, thisSrc, thisPath, thisErr string) *dealPatternError {
	return &dealPatternError{
		Owner:  thisOwner,
		Src:    thisSrc,
		Path:   thisPath,
		errInf: thisErr,
	}
}

func (dp *dealPatternError) Error() string {
	return dp.errInf
}

func (dp *dealPatternError) Kv() string {
	return executor.Kv(dp)
}

// Upgrade implements the linsener.Executor interface
type Upgrade struct {
	executor.Driver
	tmpDir         string
	backFile       string
	executePattern []string
}

// NewUpgrade construct a Upgrade and return it
func NewUpgrade(ed executor.Driver) *Upgrade {
	newr := new(Upgrade)
	newr.Driver = ed
	return newr
}

// Exec performs real upgrade tasks
func (u *Upgrade) Exec(rs *executor.Result) {
	log.Slogger.Infof("Begin to [Upgrade] service：%s,%s", u.ServiceID, u.Dir)

	var err error
	defer u.deferHandleFunc(&err, rs)

	rs.Identiy = u.Identiy

	// Backup and upload
	err = u.backup()
	if err != nil {
		rs.AppendFailedStep(executor.StepBackup, err)
		return
	}
	rs.AppendSuccessStep(executor.StepBackup)

	// Create temp dir to store code
	err = u.createTempDir()
	if err != nil {
		rs.AppendFailedStep(executor.StepGetCode, err)
		return
	}
	// Download code
	codedir, err := u.GetCode()
	if err != nil {
		rs.AppendFailedStep(executor.StepGetCode, err)
		return
	}
	u.tmpDir = codedir
	rs.AppendSuccessStep(executor.StepGetCode)

	// Verify code and service
	err = u.checkenv()
	if err != nil {
		rs.AppendFailedStep(executor.StepCheckEnv, err)
		return
	}
	rs.AppendSuccessStep(executor.StepCheckEnv)

	// Perform an upgrade
	err = u.upgrade()
	if err != nil {
		rs.AppendFailedStep(executor.StepUpgrade, err)
		return
	}
	rs.AppendSuccessStep(executor.StepUpgrade)

}

// deferHandleFunc is a defer func which handles errors and some finishing touches.
// for example, clean temporary dir or rollback
func (u *Upgrade) deferHandleFunc(err *error, rs *executor.Result) {
	if *err != nil {
		rs.ReturnCode = common.ReturnCodeFailed
		rs.ReturnMsg = (*err).Error()
		// Assert the interface type(CoulsonError)
		if ce, ok := errors.Cause(*err).(executor.CoulsonError); ok {
			log.Slogger.Errorf("encounter an error:%+v, the kv is: %s", *err, ce.Kv())

			// Rollback if dealPatternDirs or dealPatternFiles fails
			if _, ok := ce.(*dealPatternError); ok {
				err1 := u.rollback()
				if err1 != nil {
					log.Slogger.Errorf("rollback faild: %s.", err1.Error())
				}
			}
		} else {
			log.Slogger.Errorf("encounter an error:%+v.", *err)
		}
	}

	// Clean temporary directory
	if afis.IsExists(u.tmpDir) {
		log.Slogger.Infof("clean temp dir %s.", u.tmpDir)
		err2 := os.RemoveAll(u.tmpDir)
		if err2 != nil {
			log.Slogger.Errorf("remove dir faild: %s.", err2.Error())
		}
	}
	log.Slogger.Infof("Exit goroutine.")
}

// Check && Verify
func (u *Upgrade) checkenv() error {

	log.Slogger.Debugf("CustomPattern and CodePattern:%+v, %d, %+v, %d", u.CustomPattern, len(u.CustomPattern), u.CodePattern, len(u.CodePattern))

	if afis.ContainsString(u.CodePattern, "") || (afis.ContainsString(u.CustomPattern, "") && len(u.CustomPattern) > 1) {
		log.Slogger.Errorf("Pattern has empty string.")
		return errors.New("pattern has empty string")
	}

	if u.CustomPattern != nil && !afis.ContainsString(u.CustomPattern, "") {
		u.executePattern = u.CustomPattern
	} else {
		u.executePattern = u.CodePattern
	}

	if u.executePattern == nil {
		return errors.New("pattern is nil")
	}
	// Check if the deployment path already exists
	if !afis.IsExists(u.Dir) {
		return errors.WithStack(executor.NewPathError(u.Dir, "program is not exist"))
	}
	// Check if the corresponding path of CodePattern exists in the deployment directory
	for _, d := range u.executePattern {
		deployPath := path.Join(u.Dir, d)
		if strings.ContainsAny(deployPath, "*?[]") {
			deployPath = filepath.Dir(deployPath)
		}
		if !afis.IsExists(deployPath) {
			return errors.WithStack(executor.NewPathError(deployPath, "deployPath is not exist"))
		}

	}

	// Check if the corresponding path of CodePattern exists in the temporary code directory.
	// return *pathError if not
	for _, d := range u.executePattern {
		tempCodePath := path.Join(u.tmpDir, u.ModuleName, d)
		if strings.ContainsAny(tempCodePath, "*?[]") {
			tempCodePath = filepath.Dir(tempCodePath)
		}
		if !afis.IsExists(tempCodePath) {
			return errors.WithStack(executor.NewPathError(tempCodePath, "patterndir is not exist"))
		}

	}
	log.Slogger.Info("checkenv  successfully.")

	return nil
}

func (u *Upgrade) createTempDir() error {
	dir, err := ioutil.TempDir("", "upg_")
	if err != nil {
		return errors.WithStack(executor.NewPathError("/tmp/upg_", err.Error()))

	}
	u.tmpDir = dir
	log.Slogger.Infof("temp dir is : %s", u.tmpDir)
	return nil
}

func (u *Upgrade) upgrade() error {
	classifiedDirs, classifiedFiles, classifiedPatterns := u.classifyPattern()
	log.Slogger.Debugw("The paths need to be precessed:", "classifiedDirs", classifiedDirs,
		"classifiedFiles", classifiedFiles, "classifiedPatterns", classifiedPatterns)

	// Process dirs
	if classifiedDirs != nil {
		err := u.dealPatternDirs(classifiedDirs)
		if err != nil {
			return err
		}
	}
	// Process files
	if classifiedFiles != nil {
		err := u.dealPatternFiles(classifiedFiles)
		if err != nil {
			return err
		}
	}
	// Process classifiedPatterns
	if classifiedPatterns != nil {
		err := u.dealPatterns(classifiedPatterns)
		if err != nil {
			return err
		}
	}

	// Change the owner of the entire folder
	err := afis.ChownDirR(u.Dir, u.OsUser)
	if err != nil {
		return errors.WithStack(executor.NewPathError(u.Dir, err.Error()))
	}

	// Change permissions for the entire folder
	err = afis.ChmodDirR(u.Dir, 0755)
	if err != nil {
		return errors.WithStack(executor.NewPathError(u.Dir, err.Error()))
	}

	return nil
}

type pairSlice []map[string]string

// classifyPattern classifies CodePattern according to directories, files and patterns
func (u *Upgrade) classifyPattern() (pairSlice, pairSlice, pairSlice) {
	var classfiedDirs pairSlice
	var classfiedFiles pairSlice
	var classfiedPatterns pairSlice

	log.Slogger.Debugf("ExecutePattern:%+v", u.executePattern)
	for _, p := range u.executePattern {
		absPath := path.Join(u.Dir, p)
		// Load both the pattern absolute path and the relative path into the map
		mapPair := make(map[string]string)
		mapPair["pattern"] = p
		mapPair["abspath"] = absPath
		if afis.IsDir(absPath) {
			classfiedDirs = append(classfiedDirs, mapPair)
			continue
		}

		if afis.IsFile(absPath) {
			classfiedFiles = append(classfiedFiles, mapPair)
			continue
		}
		// Load the path that is neither a directory nor a file (real pattern) into the classfiedPatterns
		classfiedPatterns = append(classfiedPatterns, mapPair)
	}
	return classfiedDirs, classfiedFiles, classfiedPatterns
}

// Process directories in patterns
func (u *Upgrade) dealPatternDirs(classfiedDirs pairSlice) error {
	for _, dirMap := range classfiedDirs {
		// Check if the owner matches
		if !afis.CheckFileOwner(dirMap["abspath"], u.OsUser) {
			return errors.WithStack(executor.NewFileOwnerError(dirMap["abspath"], u.OsUser, "file and owner does not match"))
		}
		err := afis.RemoveContents(dirMap["abspath"])
		if err != nil {
			return errors.WithStack(newDealPatternError(u.OsUser, "", dirMap["abspath"], err.Error()))
		}
		// Copy direactory
		src := filepath.Join(u.tmpDir, u.ModuleName, dirMap["pattern"])
		err = afis.CopyDir(src, dirMap["abspath"])
		if err != nil {
			return errors.WithStack(newDealPatternError(u.OsUser, src, dirMap["abspath"], err.Error()))
		}
	}
	return nil
}

// Process files in patterns
func (u *Upgrade) dealPatternFiles(classfiedFiles pairSlice) error {
	for _, fileMap := range classfiedFiles {
		// Check if the owner matches
		if !afis.CheckFileOwner(fileMap["abspath"], u.OsUser) {
			return errors.WithStack(executor.NewFileOwnerError(fileMap["abspath"], u.OsUser, "file and owner does not match"))
		}
		err := os.Remove(fileMap["abspath"])
		if err != nil {
			return errors.WithStack(newDealPatternError(u.OsUser, "", fileMap["abspath"], err.Error()))
		}
		// Copy file
		src := filepath.Join(u.tmpDir, u.ModuleName, fileMap["pattern"])
		err = afis.CopyFile(src, fileMap["abspath"])
		if err != nil {
			return errors.WithStack(newDealPatternError(u.OsUser, src, fileMap["abspath"], err.Error()))
		}
	}
	return nil
}

// Process patterns
func (u *Upgrade) dealPatterns(classfiedPatterns pairSlice) error {
	var classfiedDirs pairSlice
	var classfiedFiles pairSlice
	for _, patternMap := range classfiedPatterns {
		baseDir := filepath.Dir(patternMap["abspath"])
		// Only walk current pattern's root dir for once, and process according to file and directory
		err := afis.WalkOnce(baseDir, func(path string, info os.FileInfo, err1 error) error {
			if err1 != nil {
				return errors.WithStack(newDealPatternError(u.OsUser, "", baseDir, err1.Error()))
			}
			// Check whether the path matches the pattern
			match, err2 := filepath.Match(patternMap["abspath"], path)

			if err2 != nil {
				return errors.WithStack(newDealPatternError(u.OsUser, patternMap["abspath"], path, err2.Error()))
			}
			// Classify according to directories, files
			if match {
				mapPair := make(map[string]string)
				mapPair["pattern"] = filepath.Join(filepath.Dir(patternMap["pattern"]), filepath.Base(path))
				mapPair["abspath"] = path

				if afis.IsDir(path) {
					classfiedDirs = append(classfiedDirs, mapPair)
				}

				if afis.IsFile(path) {
					classfiedFiles = append(classfiedFiles, mapPair)
				}
			}

			return nil
		})

		if err != nil {
			return err
		}
	}
	log.Slogger.Debugw("The paths have been matched:", "dirs", classfiedDirs,
		"files", classfiedFiles)
	// Process the matched directories
	err := u.dealPatternDirs(classfiedDirs)
	if err != nil {
		return err
	}
	// Process the matched files
	err = u.dealPatternFiles(classfiedFiles)
	if err != nil {
		return err
	}
	return nil
}

// Backup and upload
func (u *Upgrade) backup() error {
	// build filename
	filename := filepath.Base(u.Dir) + time.Now().Format("20060102150405.00000") + ".zip"
	// build dst file path
	dst := filepath.Join(common.TempBackupPath, filename)
	// build upload path
	upath := filepath.Join(common.AgentID, u.ServiceID)
	// backup and upload
	err := u.BackupService(dst, upath)

	if err != nil {
		return err
	}
	// build filepath with name that reside on file server
	remoteFilePath := filepath.Join(upath, filename)
	// local version file which contains remoteFilePath
	versionFile := filepath.Join(u.Dir, common.PathFile)
	err = ioutil.WriteFile(versionFile, []byte(remoteFilePath), 0644)

	if err != nil {
		return errors.WithStack(err)
	}
	// change owner
	err = afis.ChownFile(versionFile, u.OsUser)

	if err != nil {
		return errors.WithStack(err)
	}
	// record temporary backup file for rolling back
	u.backFile = dst
	return nil
}

// Rollback
func (u *Upgrade) rollback() error {
	// Check if the owner matches
	if !afis.CheckFileOwner(u.Dir, u.OsUser) {
		return errors.WithStack(executor.NewFileOwnerError(u.Dir, u.OsUser, "file and owner does not match"))
	}
	err := os.RemoveAll(u.Dir)
	if err != nil {
		return errors.WithStack(err)
	}
	// unzip temporary backup file to basedir
	basedir := filepath.Dir(u.Dir)
	err = afis.Unzip(u.backFile, basedir)
	if err != nil {
		return err
	}
	err = afis.ChownDirR(u.Dir, u.OsUser)
	if err != nil {
		return err
	}
	return nil
}
