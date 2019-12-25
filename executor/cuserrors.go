package executor

import "encoding/json"

const (
	StepCheckEnv string = "checkenv"
	StepCreateUser		= "CreateUser"
	StepCreateTmpDir	= "CreateTmpDir"
	StepGetCode         = "getcode"
	StepDeploy          = "deploy"
	StepBackup          = "backup"
	StepUpgrade         = "upgrade"
	StepRoll            = "rollback"
	StepCheck           = "check"
	StepGetPid          = "getpid"
	StepStart           = "start"
	StepStop            = "stop"
	StepRegister		= "register"
	StepDelete			= "delete"
)

type CoulsonError interface {
	error
	Kv() string
}

type pathError struct {
	Path   string `json:"path"`
	errInf string
}

func Kv(ceStruct CoulsonError) string {
	jsonByte, err := json.Marshal(ceStruct)
	if err != nil {
		return err.Error()
	}
	return string(jsonByte)
}

func NewPathError(thisPath, thisErr string) *pathError {
	return &pathError{
		Path:   thisPath,
		errInf: thisErr,
	}
}

func (p *pathError) Error() string {
	return p.errInf
}

func (p *pathError) Kv() string {
	return Kv(p)
}

type getCodeError struct {
	Url    string `json:"url"`
	errInf string
}

func NewGetCodeError(thisUrl, thisErr string) *getCodeError {
	return &getCodeError{
		Url:    thisUrl,
		errInf: thisErr,
	}
}

func (gc *getCodeError) Error() string {
	return gc.errInf
}

func (gc *getCodeError) Kv() string {
	return Kv(gc)
}

type fileOwnerError struct {
	File   string `json:"file"`
	Owner  string `json:"owner"`
	errInf string
}

func NewFileOwnerError(thisfile, thisOwner, thisErr string) *fileOwnerError {
	return &fileOwnerError{
		File:   thisfile,
		Owner:  thisOwner,
		errInf: thisErr,
	}
}

func (fo *fileOwnerError) Error() string {
	return fo.errInf
}

func (fo *fileOwnerError) Kv() string {
	return Kv(fo)
}

type cmdError struct {
	CMD    string `json:"cmd"`
	errInf string
}

func NewCmdError(thisCMD, thisErr string) *cmdError {
	return &cmdError{
		CMD:    thisCMD,
		errInf: thisErr,
	}
}

func (ce *cmdError) Error() string {
	return ce.errInf
}

func (ce *cmdError) Kv() string {
	return Kv(ce)
}
