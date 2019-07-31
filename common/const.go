/**
* @Author: xhzhang
* @Date: 2019-04-19 16:57
 */
package common

type OpMode int32

const (
	Operate_DEP OpMode = 1 //deploy
	Operate_UPG OpMode = 2 //upgrade
	Operate_STA OpMode = 3 //start
	Operate_SHU OpMode = 4 //stop
	Operate_RES OpMode = 5 //restart
	Operate_CHE OpMode = 6 //check
	Operate_BAK OpMode = 7 //backup
	Operate_ROL OpMode = 8 //rollback
)

type ExecuteReturnCode int32

const (
	ReturnCode_FAILED    ExecuteReturnCode = 0 //执行失败
	ReturnCode_SUCCESS   ExecuteReturnCode = 1 //执行成功
	ReturnCode_Exception ExecuteReturnCode = 2 //执行异常
)

const (
	ReturnOKMsg string = "OK"
	ConfigTemplate = "/agentConfig/template"
	UUIDFile = "/etc/agent/uuid"
	PathFile = ".version"
)
