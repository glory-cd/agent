/**
* @Author: xhzhang
* @Date: 2019-04-18 16:36
 */
package core

import (
	"github.com/glory-cd/agent/executor"
	"github.com/glory-cd/utils/afis"
	"github.com/pkg/errors"
	"log"
	"strings"
)

type Agent struct {
	AgentID            string
	EtcdKey            string
	EtcdVal            string
	InstructionChannel string
	GraceChannel       string
	Services           []executor.Service
	ServicePrefix      string
}

func NewAgent() *Agent {
	var agent = new(Agent)
	return agent
}

func (a *Agent) SetAgentID(uuidfile string) {
	if !afis.IsExists(uuidfile) {
		err := afis.WriteUUID2File(uuidfile) //文件不存在，则重新生成uuid写入文件
		if err != nil {
			log.Fatalf("write uuid to [%s] failed: %+v", uuidfile, errors.WithStack(err))
		}
	}
	uuid, err := afis.ReadUUIDFromFile(uuidfile)
	if err != nil {
		log.Fatalf("read uuidfile [%s] failed: %+v", uuidfile, errors.WithStack(err))
	}

	a.AgentID = uuid
}

func (a *Agent) SetEtcdKey() {
	a.EtcdKey = "/agent/" + a.AgentID
}

func (a *Agent) SetEtcdVal() {
	localip, err := afis.GetLocalIP()
	if err != nil {
		log.Fatalf("GetIP Err:%+v", errors.WithStack(err))
	}
	hostname, err := afis.GetHostName()
	if err != nil {
		log.Fatalf("GetHostName Err:%+v", errors.WithStack(err))
	}
	a.EtcdVal = hostname + ":" + strings.Join(localip, ";")
}

func (a *Agent) SetInstructionChannel() {
	a.InstructionChannel = "cmd." + a.AgentID
	a.GraceChannel = "grace." + a.AgentID
}

// set []Service to a.Services
func (a *Agent) SetServicesStruct(slist []executor.Service) {
	a.Services = slist
}

//set config template
func (a *Agent) SetServicePrefix() {
	a.ServicePrefix = "/service/" + a.AgentID + "/"
}

/*
	根据serviceid返回服务
*/
func (a *Agent) GetService(serviceid string) executor.Service {
	var ser executor.Service
	for _, s := range a.Services {
		if serviceid == s.ServiceID {
			ser = s
		}
	}

	return ser
}

//返回存在与不存在的服务slice

func (a *Agent) CheckServiceIsExist(sidlist []string) ([]string, []string) {
	var existServices, notExistServices []string

	flag := make(map[string]bool)
	for _, v := range a.Services {
		flag[v.ServiceID] = true
	}

	for _, sid := range sidlist {
		if flag[sid] {
			existServices = append(existServices, sid)
		} else {
			notExistServices = append(existServices, sid)
		}
	}

	return existServices, notExistServices

}

// 服务注册时，用于检查该服务是否存在
func (a *Agent) CheckRegisterIsExist(id string) bool {
	for _, v := range a.Services {
		if v.ServiceID == id {
			return true
		}
	}
	return false
}

//增加新服务
func (a *Agent) AddService(s executor.Service) {
	a.Services = append(a.Services, s)
}

//同步服务
func (a *Agent) SyncService(s executor.Service) {
	for index, v := range a.Services {
		if v.ServiceID == s.ServiceID {
			//先删除再添加，不过slice不适合做删除动作，后续考虑使用container/list
			a.Services = append(a.Services[:index], a.Services[index+1:]...)
			a.Services = append(a.Services, s)
		}
	}
}
