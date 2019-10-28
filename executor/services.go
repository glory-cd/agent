/**
* @Author: xhzhang
* @Date: 2019-04-25 13:48
 */
package executor

import (
	"bytes"
	"encoding/json"
	"github.com/glory-cd/agent/common"
	"github.com/glory-cd/utils/afis"
)

// Build a Service with ServiceID from json string
func NewServiceFromJson(sjson string) (Service, error) {
	var s Service
	err := json.Unmarshal([]byte(sjson), &s)
	if err != nil {
		return s, err
	}
	s.ServiceID = afis.GetMd5String(common.AgentID + s.Dir)
	return s, nil
}

// convert Service to json string
func NewJsonFromService(s Service) (string, error) {
	jsonb, err := json.Marshal(s)
	sjson := bytes.NewBuffer(jsonb).String()
	if err != nil {
		return "", err
	}
	return sjson, nil
}
