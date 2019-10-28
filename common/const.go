/**
* @Author: xhzhang
* @Date: 2019-04-19 16:57
 */
package common

type OpMode int32

const (
	OperateDEP OpMode = 1 //deploy
	OperateUPG OpMode = 2 //upgrade
	OperateSTA OpMode = 3 //start
	OperateSHU OpMode = 4 //stop
	OperateRES OpMode = 5 //restart
	OperateCHE OpMode = 6 //check
	OperateBAK OpMode = 7 //backup
	OperateROL OpMode = 8 //rollback
)

type ExecuteReturnCode int32

const (
	ReturnCodeFailed    ExecuteReturnCode = 0
	ReturnCodeSuccess   ExecuteReturnCode = 1
	ReturnCodeException ExecuteReturnCode = 2
)

const (
	ReturnOKMsg string = "OK"
	ConfigTemplate string = "/agentConfig/template"
	UUIDFile string = "/etc/agent/uuid"
	PathFile string = ".version"
	TempBackupPath string = "/tmp/backup"
	RegisterScript string = "meta.sh"
)
