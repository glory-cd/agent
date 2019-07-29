/**
* @Author: xhzhang
* @Date: 2019-04-25 13:48
 */
package executor

import (
	"bytes"
	"encoding/json"
	"github.com/auto-cdp/agent/common"
	"github.com/auto-cdp/utils/afis"
)

func NewServiceFromJson(sjson string) (Service, error) {
	var s Service
	err := json.Unmarshal([]byte(sjson), &s)
	if err != nil {
		return s, err
	}
	//log.Slogger.Debugw("generate service id.", "agentid", common.AgentID, "dir", s.Dir)
	s.ServiceID = afis.GetMd5String(common.AgentID + s.Dir)
	return s, nil
}

func NewJsonFromService(s Service) (string, error) {
	jsonb, err := json.Marshal(s)
	sjson := bytes.NewBuffer(jsonb).String()
	if err != nil {
		return "", err
	}
	return sjson, nil
}
