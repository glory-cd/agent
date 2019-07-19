/**
* @Author: xhzhang
* @Date: 2019-04-19 16:57
 */
package common

type OpMode int32

const (
	Operate_DEP OpMode = 0 //deploy
	Operate_UPG OpMode = 1 //upgrade
	Operate_STA OpMode = 2 //start
	Operate_SHU OpMode = 3 //stop
	Operate_RES OpMode = 4 //restart
	Operate_CHE OpMode = 5 //check
)

type FileOpMode int32

const (
	FileOp_ADD  FileOpMode = 0 //添加字段
	FileOp_DEL  FileOpMode = 1 //删除字段
	FileOp_EDIT FileOpMode = 2 //修改字段
)

type ExecuteReturnCode int32

const (
	ReturnCode_FAILED    ExecuteReturnCode = 0 //执行失败
	ReturnCode_SUCCESS   ExecuteReturnCode = 1 //执行成功
	ReturnCode_Exception ExecuteReturnCode = 2 //执行异常
)

type ServiceExistCode int32

const (
	ServiceNonExist      ServiceExistCode = 0
	ServiceExistAndMatch ServiceExistCode = 1
	ServiceExistNotMatch ServiceExistCode = 2
)

const (
	ReturnOKMsg string = "OK"
	ConfigTemplate = "/agentConfig/template"
	UUIDFile = "/etc/agent/uuid"
)
