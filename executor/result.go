/**
* @Author: xhzhang
* @Date: 2019-04-25 11:13
 */
package executor

import (
	"agent/common"
	"encoding/json"
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


func NewResultPointer(id Identiy) *Result {
	var r = new(Result)
	r.Identiy = id
	r.ReturnCode = common.ReturnCode_SUCCESS
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


//构造StepInfo
func (r *Result) AppendResultStep(stepnum int, stepname string, stepstate common.ExecuteReturnCode, stepmsg string, rtime int64) {
	s := ResultStep{stepnum, stepname, stepstate, stepmsg, rtime}
	r.StepInfo = append(r.StepInfo, s)
}
