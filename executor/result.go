/**
* @Author: xhzhang
* @Date: 2019-04-25 11:13
 */
package executor

import (
	"encoding/json"
	"github.com/glory-cd/agent/common"
	"time"
)

type Result struct {
	Identiy
	ReturnCode common.ExecuteReturnCode `json:"rcode"`
	ReturnMsg  string                   `json:"rmsg"`
	StepInfo   []ResultStep             `json:"rsteps"`
}

type ResultStep struct {
	StepNum    int                      `json:"stepnum"`
	StepName   string                   `json:"stepname"`
	StepState  common.ExecuteReturnCode `json:"stepstate"`
	StepMsg    string                   `json:"stepmsg"`
	ResultTime int64                    `json:"steptime"`
}

func NewResult() *Result {
	var r = new(Result)
	r.ReturnCode = common.ReturnCodeSuccess
	r.ReturnMsg = common.ReturnOKMsg
	return r
}

func (r *Result) ToJsonString() (string, error) {
	resultbyte, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(resultbyte), nil

}

// Build StepInfo
func (r *Result) AppendFailedStep(stepname string, err error) {
	stepstate := common.ReturnCodeFailed
	stepmsg := err.Error()

	stepnum := len(r.StepInfo) + 1

	s := ResultStep{stepnum, stepname, stepstate, stepmsg, time.Now().UnixNano()}

	r.StepInfo = append(r.StepInfo, s)
}

func (r *Result) AppendSuccessStep(stepname string) {
	stepstate := common.ReturnCodeSuccess
	stepmsg := common.ReturnOKMsg
	stepnum := len(r.StepInfo) + 1

	s := ResultStep{stepnum, stepname, stepstate, stepmsg, time.Now().UnixNano()}

	r.StepInfo = append(r.StepInfo, s)
}
