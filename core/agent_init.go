/**
* @Author: xhzhang
* @Date: 2019-04-19 10:12
 */
package core

import (
	"agent/common"
	"agent/executor"
	"strings"
	"utils/log"
)

var CurAgent *Agent

func InitAgent(etc string) {
	//初始化agent
	CurAgent = NewAgent()
	CurAgent.SetAgentID(common.UUIDFile)
	CurAgent.SetEtcdKey()
	CurAgent.SetEtcdVal()
	CurAgent.SetServicePrefix()
	CurAgent.SetInstructionChannel()
	//设置etcd地址
	common.EtcdEndpoint = strings.Split(etc, ",")
	//设置AgentID
	common.AgentID = CurAgent.AgentID
	//设置当前agent在etcd中的ConfigKey
	common.ConfigKey = "/agentConfig/" + common.AgentID
	//初始化etcd client
	common.InitEtcdClient()
	//初始化配置
	common.InitConfig()
	//初始化日志
	common.InitLog()
	//初始化redis
	common.InitRedis()
	//从etcd中获取属于当前agent的服务
	localServices, err := getServicesFromEtcd()
	if err != nil {
		log.Slogger.Fatalf("Agent init failed =>getServicesFromEtcd Err:[%s]", err)
	} else {
		//在Agent中设置服务
		CurAgent.SetServicesStruct(localServices)
	}

}

//从etcd中获取属于当前agent的服务
func getServicesFromEtcd() ([]executor.Service, error) {
	var servicelist []executor.Service
	servicesSlice, err := common.EtcdClient.GetWithPrefix(CurAgent.ServicePrefix)
	if err != nil {
		return nil, err
	}
	for _, ser := range servicesSlice {
		serviceStruct, err := executor.NewServiceFromJson(ser)
		if err != nil {
			return nil, err
		}
		servicelist = append(servicelist, serviceStruct)
	}
	return servicelist, nil
}
