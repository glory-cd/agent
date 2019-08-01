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
	rs       Result
	tmpdir   string
	backfile string
}

func (u *Upgrade) Exec(out chan<- Result) {
	log.Slogger.Infof("开始升级服务：%s,%s", u.ServiceID, u.Dir)

	var err error
	defer u.deferHandleFunc(&err, out)

	//执行备份
	err = u.backup()
	if err != nil {
		u.constructRS(upgradeStepCodeBackup, common.ReturnCode_FAILED, err.Error())
		return
	}

	u.constructRS(upgradeStepCodeBackup, common.ReturnCode_SUCCESS, common.ReturnOKMsg)

	//创建临时代码目录
	err = u.createTempDir()
	if err != nil {
		u.constructRS(upgradeStepCodeGetCode, common.ReturnCode_FAILED, err.Error())
		return
	}
	//下载代码
	err = u.getCode()
	if err != nil {
		u.constructRS(upgradeStepCodeGetCode, common.ReturnCode_FAILED, err.Error())
		return
	}
	u.constructRS(upgradeStepCodeGetCode, common.ReturnCode_SUCCESS, common.ReturnOKMsg)

	//检查代码以及service
	err = u.checkenv()
	if err != nil {
		u.constructRS(upgradeStepCodeCheck, common.ReturnCode_FAILED, err.Error())
		return
	}
	u.constructRS(upgradeStepCodeCheck, common.ReturnCode_SUCCESS, common.ReturnOKMsg)

	//执行升级
	err = u.upgrade()
	if err != nil {
		u.constructRS(upgradeStepCodeUpgrade, common.ReturnCode_FAILED, err.Error())
		return
	}
	u.constructRS(upgradeStepCodeUpgrade, common.ReturnCode_SUCCESS, common.ReturnOKMsg)

}

func (u *Upgrade) deferHandleFunc(err *error, out chan<- Result) {
	if *err != nil {
		u.rs.ReturnCode = common.ReturnCode_FAILED
		u.rs.ReturnMsg = (*err).Error()
		//断言err的接口类型为CoulsonError
		if ce, ok := errors.Cause(*err).(CoulsonError); ok {
			log.Slogger.Errorf("encounter an error:%+v, the kv is: %s", *err, ce.Kv())

			//如果dealPatternDirs和dealPatternFiles失败，则执行回滚
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

	//清理临时目录
	if afis.IsExists(u.tmpdir) {
		log.Slogger.Infof("clean temp dir %s.", u.tmpdir)
		err2 := os.RemoveAll(u.tmpdir)
		if err2 != nil {
			log.Slogger.Errorf("remove dir faild: %s.", err2.Error())
		}
	}
	//结果写入chanel
	out <- u.rs
	log.Slogger.Infof("退出goroutine.")
}

//构造ResultService
func (u *Upgrade) constructRS(step int, rcode common.ExecuteReturnCode, errstr string) {
	u.rs.AppendResultStep(
		step,
		upgradeStepName[step],
		rcode,
		errstr,
		time.Now().UnixNano(),
	)
}

//检查环境
func (u *Upgrade) checkenv() error {
	var executePattern []string
	if u.CustomPattern != nil {
		executePattern = u.CustomPattern
	} else {
		executePattern = u.CodePattern
	}
	//检查部署路径是否已存在，若不存在则返回错误
	if !afis.IsExists(u.Dir) {
		return errors.WithStack(NewPathError(u.Dir, "program is not exist"))
	}
	//检查部署目录中CodePattern的路径是否存在, 如果不存在则仅记录异常
	for _, dcp := range executePattern {
		codepath := path.Join(u.Dir, dcp)
		if strings.ContainsAny(codepath, "*?[]") {
			codepath = filepath.Dir(codepath)
		}
		if !afis.IsExists(codepath) {
			return errors.WithStack(NewPathError(codepath, "codepath is not exist"))
		}

	}

	//检查临时代码目录中CodePattern的路径是否存在，如果不存在则返回*pathError
	for _, tcp := range executePattern {
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

func (u *Upgrade) getCode() error {
	//从url获取代码
	log.Slogger.Debugf("download code from url: %s", u.RemoteCode)
	err := afis.DownloadCode(u.tmpdir, u.RemoteCode)
	if err != nil {
		return errors.WithStack(NewGetCodeError(u.RemoteCode, err.Error()))
	}
	log.Slogger.Infof("download code to %s", u.tmpdir)
	return nil
}

func (u *Upgrade) upgrade() error {
	patterndirs, patternfiles, patterns := u.classifyPattern()
	log.Slogger.Debugw("The paths need to be precessed:", "patterndirs", patterndirs,
		"patternfiles", patternfiles, "patterns", patterns)

	if patterndirs != nil {
		err := u.dealPatternDirs(patterndirs)
		if err != nil {
			return err
		}
	}

	if patternfiles != nil {
		err := u.dealPatternFiles(patternfiles)
		if err != nil {
			return err
		}
	}

	if patterns != nil {
		err := u.dealPatterns(patterns)
		if err != nil {
			return err
		}
	}

	//更改整个文件夹的属主
	err := afis.ChownDirR(u.Dir, u.OsUser)
	if err != nil {
		return errors.WithStack(NewPathError(u.Dir, err.Error()))
	}

	return nil
}

//将CodePattern按具体目录、具体文件、Pattern进行分类
func (u *Upgrade) classifyPattern() (patterndirs, patternfiles, patterns []map[string]string) {
	var pdirs []map[string]string
	var pfiles []map[string]string
	var ppatterns []map[string]string
	var executePattern []string
	log.Slogger.Debugf("custom:%+v, %d", u.CustomPattern, len(u.CustomPattern))
	if u.CustomPattern != nil && len(u.CustomPattern) != 0 {
		executePattern = u.CustomPattern
	} else {
		executePattern = u.CodePattern
	}
	for _, cp := range executePattern {
		codepath := path.Join(u.Dir, cp)
		//将pattern绝对路径和相对路径都装入map
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
		//对于既不是目录也不是文件（真正的pattern）的路径装入ppatterns
		ppatterns = append(ppatterns, dm)
	}
	return pdirs, pfiles, ppatterns
}

//处理patterns中的目录
func (u *Upgrade) dealPatternDirs(pdirs []map[string]string) error {
	for _, pd := range pdirs {
		//如果要删除内容的目录属主与服务所在用户不同则直接返回*depErrors
		if !afis.CheckFileOwner(pd["codepath"], u.OsUser) {
			return errors.WithStack(
				NewFileOwnerError(pd["codepath"],
					u.OsUser,
					"file and owner does not match"))
		}
		err := afis.RemoveContents(pd["codepath"])
		if err != nil {
			return errors.Wrap(
				NewdealPatternError(
					u.OsUser,
					"",
					pd["codepath"],
					err.Error(),
				),
				"upgrade.dealPatternDirs.afis.RemoveContents",
			)
		}
		src := filepath.Join(u.tmpdir, u.ModuleName, pd["codepattern"])
		err = afis.CopyDir(src, pd["codepath"])
		if err != nil {
			return errors.Wrap(
				NewdealPatternError(
					u.OsUser,
					src,
					pd["codepath"],
					err.Error(),
				),
				"upgrade.dealPatternDirs.afis.CopyDir",
			)
		}
	}
	return nil
}

//处理patterns中的文件
func (u *Upgrade) dealPatternFiles(pfiles []map[string]string) error {
	for _, pf := range pfiles {
		//如果要删除的文件属主与服务所在用户不同则直接返回×depErrors
		if !afis.CheckFileOwner(pf["codepath"], u.OsUser) {
			return errors.WithStack(
				NewFileOwnerError(pf["codepath"],
					u.OsUser,
					"file and owner does not match"))
		}
		err := os.Remove(pf["codepath"])
		if err != nil {
			return errors.Wrap(
				NewdealPatternError(
					u.OsUser,
					"",
					pf["codepath"],
					err.Error(),
				),
				"upgrade.dealPatternFiles.os.Remove",
			)
		}
		src := filepath.Join(u.tmpdir, u.ModuleName, pf["codepattern"])
		err = afis.CopyFile(src, pf["codepath"])
		if err != nil {
			return errors.Wrap(
				NewdealPatternError(
					u.OsUser,
					src,
					pf["codepath"],
					err.Error(),
				),
				"upgrade.dealPatternFiles.afis.CopyFile",
			)
		}
	}
	return nil
}

//处理patterns中的pattern
func (u *Upgrade) dealPatterns(ppatterns []map[string]string) error {
	var pdirs []map[string]string
	var pfiles []map[string]string
	for _, pp := range ppatterns {
		basedir := filepath.Dir(pp["codepath"])
		//以pattern的基目录为起点进行一级目录遍历
		err := afis.WalkOnce(basedir, func(path string, info os.FileInfo, err1 error) error {
			if err1 != nil {
				return errors.Wrap(
					NewdealPatternError(
						u.OsUser,
						"",
						basedir,
						err1.Error(),
					),
					"upgrade.dealPatterns.afis.WalkOnce",
				)
			}
			//将遍历到的路径与pattern进行匹配
			ism, err2 := filepath.Match(pp["codepath"], path)

			if err2 != nil {
				return errors.Wrap(
					NewdealPatternError(
						u.OsUser,
						pp["codepath"],
						path,
						err2.Error(),
					),
					"upgrade.dealPatterns.filepath.Match",
				)
			}
			//如果path匹配到pattern，则将路径按目录和文件分类组合，以备调用不同的处理方法
			if ism {
				//组合为具体的目录和文件
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
	//处理匹配到的目录
	err := u.dealPatternDirs(pdirs)
	if err != nil {
		return err
	}
	//处理匹配到的文件
	err = u.dealPatternFiles(pfiles)
	if err != nil {
		return err
	}
	return nil
}

//备份
func (u *Upgrade) backup() error {

	filename := filepath.Base(u.Dir) + time.Now().Format("20060102150405.00000") + ".zip"

	dst := filepath.Join("/tmp/backup", filename)

	upath := filepath.Join(common.AgentID, u.ServiceID)

	err := u.backupService(dst, upath)

	if err != nil{
		return err
	}

	remoteFilePath := filepath.Join(upath, filename)

	versionFile := filepath.Join(u.Dir, common.PathFile)

	err = ioutil.WriteFile(versionFile, []byte(remoteFilePath), 0644)

	if err != nil {
		return errors.WithStack(err)
	}

	err = afis.ChownFile(versionFile, u.OsUser)

	if err != nil {
		return errors.WithStack(err)
	}

	u.backfile = dst
	return nil
}

//回滚
func (u *Upgrade) rollBack() error {
	//如果要删除的文件属主与服务所在用户不同则直接返回*depErrors
	if !afis.CheckFileOwner(u.Dir, u.OsUser) {
		return errors.WithStack(
			NewFileOwnerError(u.Dir,
				u.OsUser,
				"file and owner does not match"))
	}
	err := os.RemoveAll(u.Dir)
	if err != nil {
		return errors.Wrap(err, "upgrade.rollBack.RemoveAll")
	}
	basedir := filepath.Dir(u.Dir)
	err = afis.Unzip(u.backfile, basedir)
	if err != nil {
		return errors.Wrap(err, "upgrade.rollBack.afis.Unzip")
	}
	err = afis.ChownDirR(u.Dir, u.OsUser)
	if err != nil {
		return err
	}
	return nil
}
