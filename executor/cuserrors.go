package executor

import "encoding/json"

const (
	deployStepCodeCheckEnv int = iota + 1
	deployStepCodeInitEnv
	deployStepCodeGetCode
	deployStepCodeDeploy
)

var deployStepName = map[int]string{
	deployStepCodeCheckEnv: "checkenv",
	deployStepCodeInitEnv:  "initenv",
	deployStepCodeGetCode:  "getcode",
	deployStepCodeDeploy:   "deploy",
}

const (
	upgradeStepCodeBackup int = iota + 1
	upgradeStepCodeGetCode
	upgradeStepCodeCheck
	upgradeStepCodeUpgrade
)

var upgradeStepName = map[int]string{
	upgradeStepCodeBackup:  "backup",
	upgradeStepCodeGetCode: "getcode",
	upgradeStepCodeCheck:   "checkenv",
	upgradeStepCodeUpgrade: "upgrade",
}

type CoulsonError interface {
	error
	Kv() string
}

type pathError struct {
	Path   string `json:"path"`
	errInf string
}


func kv(ceStruct CoulsonError) string {
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
	return kv(p)
}

type dealPatternError struct {
	Owner  string `json:"owner"`
	Src    string `json:"src"`
	Path   string `json:"path"`
	errInf string
}

func NewdealPatternError(thisOwner, thisSrc, thisPath, thisErr string) *dealPatternError {
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
	return kv(dp)
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
	return kv(gc)
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
	return kv(fo)
}

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
	return kv(de)
}

type cmdError struct {
	CMD    string `json:"cmd"`
	errInf string
}

func NewCmdError(thisCMD, thisErr string) *cmdError {
	return &cmdError{
		CMD:        thisCMD,
		errInf:     thisErr,
	}
}

func (ce *cmdError) Error() string {
	return ce.errInf
}

func (ce *cmdError) Kv() string {
	return kv(ce)
}