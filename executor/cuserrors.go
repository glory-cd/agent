package executor

import "encoding/json"

// string const for operation step name
const (
	StepCheckEnv     string = "checkenv"
	StepCreateUser          = "CreateUser"
	StepCreateTmpDir        = "CreateTmpDir"
	StepGetCode             = "getcode"
	StepDeploy              = "deploy"
	StepBackup              = "backup"
	StepUpgrade             = "upgrade"
	StepRoll                = "rollback"
	StepCheck               = "check"
	StepGetPid              = "getpid"
	StepStart               = "start"
	StepStop                = "stop"
	StepRegister            = "register"
	StepDelete              = "delete"
)

// CoulsonError is an error interface which includes Kv() func
type CoulsonError interface {
	error
	Kv() string
}

// PathError is an error about path
type PathError struct {
	Path   string `json:"path"`
	errInf string
}

// Kv encodes a CoulsonError
func Kv(ceStruct CoulsonError) string {
	jsonByte, err := json.Marshal(ceStruct)
	if err != nil {
		return err.Error()
	}
	return string(jsonByte)
}

// NewPathError returns a *pathError
func NewPathError(thisPath, thisErr string) *PathError {
	return &PathError{
		Path:   thisPath,
		errInf: thisErr,
	}
}

func (p *PathError) Error() string {
	return p.errInf
}

// Kv implements CoulsonError
func (p *PathError) Kv() string {
	return Kv(p)
}

// GetCodeError is error about download
type GetCodeError struct {
	URL    string `json:"url"`
	errInf string
}

// NewGetCodeError returns a *getCodeError
func NewGetCodeError(thisURL, thisErr string) *GetCodeError {
	return &GetCodeError{
		URL:    thisURL,
		errInf: thisErr,
	}
}

func (gc *GetCodeError) Error() string {
	return gc.errInf
}

// Kv implements CoulsonError
func (gc *GetCodeError) Kv() string {
	return Kv(gc)
}

// FileOwnerError is an error about owner
type FileOwnerError struct {
	File   string `json:"file"`
	Owner  string `json:"owner"`
	errInf string
}

// NewFileOwnerError returns a *fileOwnerError
func NewFileOwnerError(thisfile, thisOwner, thisErr string) *FileOwnerError {
	return &FileOwnerError{
		File:   thisfile,
		Owner:  thisOwner,
		errInf: thisErr,
	}
}

func (fo *FileOwnerError) Error() string {
	return fo.errInf
}

// Kv implements CoulsonError
func (fo *FileOwnerError) Kv() string {
	return Kv(fo)
}

// CMDError is an error abut running command
type CMDError struct {
	CMD    string `json:"cmd"`
	errInf string
}

// NewCMDError returns a *CMDError
func NewCMDError(thisCMD, thisErr string) *CMDError {
	return &CMDError{
		CMD:    thisCMD,
		errInf: thisErr,
	}
}

func (ce *CMDError) Error() string {
	return ce.errInf
}

// Kv implements CoulsonError
func (ce *CMDError) Kv() string {
	return Kv(ce)
}
