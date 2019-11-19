/**
* @Author: xhzhang
* @Date: 2019-04-18 16:36
 */
package listener

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
	// If the file does not exist, the generated uuid is written to the file
	if !afis.IsExists(uuidfile) {
		err := afis.WriteUUID2File(uuidfile)
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
	Returns the service according to serviceid
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

// Returns slice with and without a service

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

// When registering a service, check if the service exists
func (a *Agent) CheckRegisterIsExist(id string) bool {
	for _, v := range a.Services {
		if v.ServiceID == id {
			return true
		}
	}
	return false
}

// Add new service
func (a *Agent) AddService(s executor.Service) {
	a.Services = append(a.Services, s)
}

// sync service
func (a *Agent) SyncService(s executor.Service) {
	for index, v := range a.Services {
		if v.ServiceID == s.ServiceID {
			// Delete before adding, but slice is not a good choice for deleting.
			// Consider container/list later
			a.Services = append(a.Services[:index], a.Services[index+1:]...)
			a.Services = append(a.Services, s)
		}
	}
}
