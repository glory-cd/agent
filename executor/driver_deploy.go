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

//部署执行
func (d *Deploy) Exec(out chan<- Result) {
	log.Slogger.Infof("开始部署服务：%s,%s", d.ServiceID, d.Dir)

	//使用defer + 闭包来处理错误返回以及清理临时代码存放目录
	var err error
	defer d.deferHandleFunc(&err, out)

	//检查环境
	err = d.checkenv()
	if err != nil {
		d.rs.AppendFailedStep(stepNameCheckEnv, err)
		return
	}
	d.rs.AppendSuccessStep(stepNameCheckEnv)

	//初始化用户目录等
	err = d.initenv(&d.rs)
	if err != nil {
		return
	}

	//下载代码
	err = d.getCode(&d.rs)
	if err != nil {
		d.rs.AppendFailedStep(stepNameGetCode, err)
		return
	}
	d.rs.AppendSuccessStep(stepNameGetCode)

	//执行部署
	err = d.deploy(&d.rs)
	if err != nil {
		d.rs.AppendFailedStep(stepNameDeploy, err)
		return
	}
	d.rs.AppendSuccessStep(stepNameDeploy)

}

func (d *Deploy) deferHandleFunc(err *error, out chan<- Result) {
	//断言err的接口类型为CoulsonError
	if *err != nil {
		d.rs.ReturnCode = common.ReturnCodeFailed
		d.rs.ReturnMsg = (*err).Error()
		if ce, ok := errors.Cause(*err).(CoulsonError); ok {
			log.Slogger.Errorf("encounter an error:%+v, the kv is: %s", *err, ce.Kv())
			//如果deploy失败，则删除创建的服务目录
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

	//清理临时目录
	if afis.IsExists(d.tempdir) {
		log.Slogger.Infof("clean temp dir %s.", d.tempdir)
		err2 := os.RemoveAll(d.tempdir)
		if err2 != nil {
			log.Slogger.Errorf("remove dir faild: %s.", err2.Error())
		}
	}
	//结果写入chanel
	out <- d.rs
	log.Slogger.Infof("退出goroutine.")
}

//获取代码并解压到临时目录
func (d *Deploy) getCode(r *Result) error {
	//从url获取代码
	fileServer := common.Config().FileServer
	dir, err := Get(fileServer, d.RemoteCode)
	//err := afis.DownloadCode(d.tempdir, d.RemoteCode)
	if err != nil {
		return err
	}
	d.tempdir = dir
	log.Slogger.Infof("download code to %s", d.tempdir)
	return nil
}

//部署
func (d *Deploy) deploy(r *Result) error {
	//创建服务目录
	err := os.Mkdir(d.Dir, 0755)
	if err != nil {
		return errors.WithStack(NewPathError(d.Dir, err.Error()))
	}

	log.Slogger.Infof("create code dir: %s", d.Dir)

	//组装路径，仅复制代码目录中的内容，不包括代码目录本身
	src := path.Join(d.tempdir, d.ModuleName)
	err = afis.CopyDir(src, d.Dir)
	if err != nil {
		return errors.Wrap(
			NewDeployError(
				src,
				d.Dir,
				err.Error(),
			),
			"deploy.deploy.afis.CopyDir",
		)
	}
	log.Slogger.Infof("copy code from %s to %s successfully.", src, d.Dir)
	//更改整个文件夹的属主
	err = afis.ChownDirR(d.Dir, d.OsUser)
	if err != nil {
		return errors.Wrap(
			NewDeployError(
				src,
				d.Dir,
				err.Error(),
			),
			"deploy.deploy.afis.ChownDirR",
		)
	}
	//更改整个文件夹的权限
	err = afis.ChmodDirR(d.Dir, 0755)
	if err != nil {
		return errors.Wrap(
			NewDeployError(
				src,
				d.Dir,
				err.Error(),
			),
			"deploy.deploy.afis.ChmodDirR",
		)
	}
	return nil
}

//检查环境
func (d *Deploy) checkenv() error {
	//检查用户是否存在
	if afis.IsUser(d.OsUser) {
		d.isuser = true
	}
	//检查部署路径是否已存在，若存在则返回错误
	if afis.IsExists(d.Dir) {
		return errors.WithStack(NewPathError(d.Dir, "deploy path already exist"))
	}
	return nil
}

func (d *Deploy) createUser() error{
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
	//处理stdout和stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	//执行
	err = cmd.Run()
	_, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	if err != nil {
		return errors.Wrap(err, errStr)
	}

	return nil
}

//初始化环境
func (d *Deploy) initenv(r *Result) error {
	//若用户不存在则创建用户
	if !d.isuser {
		err := d.createUser()
		if err != nil {
			d.rs.AppendFailedStep(stepNameCreateUser, err)
			return err
		}
		d.rs.AppendSuccessStep(stepNameCreateUser)
		log.Slogger.Infof("create user %s success!", d.Service.OsUser)
	} else {
		d.rs.AppendFailedStep(stepNameCreateUser, errors.New("The user already exists"))
	}

	//创建临时存放代码目录
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
