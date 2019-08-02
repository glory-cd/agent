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

//从json构建带有ServiceID的Service
func NewServiceFromJson(sjson string) (Service, error) {
	var s Service
	err := json.Unmarshal([]byte(sjson), &s)
	if err != nil {
		return s, err
	}
	s.ServiceID = afis.GetMd5String(common.AgentID + s.Dir)
	return s, nil
}

//将Service转为json字符串
func NewJsonFromService(s Service) (string, error) {
	jsonb, err := json.Marshal(s)
	sjson := bytes.NewBuffer(jsonb).String()
	if err != nil {
		return "", err
	}
	return sjson, nil
}
