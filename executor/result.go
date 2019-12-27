// Package executor ...
package executor

import (
	"encoding/json"
	"github.com/glory-cd/agent/common"
	"time"
)

// Result is a response message to server
// it is encoded and sent to redis
type Result struct {
	Identiy
	ReturnCode common.ExecuteReturnCode `json:"rcode"`
	ReturnMsg  string                   `json:"rmsg"`
	StepInfo   []ResultStep             `json:"rsteps"`
}

// ResultStep contains every step info
type ResultStep struct {
	StepNum    int                      `json:"stepnum"`
	StepName   string                   `json:"stepname"`
	StepState  common.ExecuteReturnCode `json:"stepstate"`
	StepMsg    string                   `json:"stepmsg"`
	ResultTime int64                    `json:"steptime"`
}

// NewResult returns a *Result
func NewResult() *Result {
	var r = new(Result)
	r.ReturnCode = common.ReturnCodeSuccess
	r.ReturnMsg = common.ReturnOKMsg
	return r
}

// ToJSONString encodes a *Result using json
func (r *Result) ToJSONString() (string, error) {
	resultbyte, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(resultbyte), nil

}

// AppendFailedStep appends failed StepInfo to *Result
func (r *Result) AppendFailedStep(stepname string, err error) {
	stepstate := common.ReturnCodeFailed
	stepmsg := err.Error()

	stepnum := len(r.StepInfo) + 1

	s := ResultStep{stepnum, stepname, stepstate, stepmsg, time.Now().UnixNano()}

	r.StepInfo = append(r.StepInfo, s)
}

// AppendSuccessStep appends sucess StepInfo to *Result
func (r *Result) AppendSuccessStep(stepname string) {
	stepstate := common.ReturnCodeSuccess
	stepmsg := common.ReturnOKMsg
	stepnum := len(r.StepInfo) + 1

	s := ResultStep{stepnum, stepname, stepstate, stepmsg, time.Now().UnixNano()}

	r.StepInfo = append(r.StepInfo, s)
}
