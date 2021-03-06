// Package listener handles instruction
package listener

import (
	"github.com/glory-cd/agent/common"
	"github.com/glory-cd/agent/executor"
	"github.com/glory-cd/utils/log"
	"strings"
)

// CurAgent is an runtime agent instance
var CurAgent *Agent

// InitAgent inits current agent info
func InitAgent(etc string) {
	// Initialize agent
	CurAgent = NewAgent()
	CurAgent.SetAgentID(common.UUIDFile)
	CurAgent.SetEtcdKey()
	CurAgent.SetEtcdVal()
	CurAgent.SetServicePrefix()
	CurAgent.SetInstructionChannel()
	// Set the etcd address
	common.EtcdEndpoint = strings.Split(etc, ",")
	// Set the AgentID
	common.AgentID = CurAgent.AgentID
	// Set ConfigKey of the current agent in etcd
	common.ConfigKey = "/agentConfig/" + common.AgentID
	// Initialize the etcd client
	common.InitEtcdClient()
	// Initialize configuration
	common.InitConfig()
	// Initialize log
	common.InitLog()
	//Initialize redis
	common.InitRedis()
	// Get the service that belongs to the current agent from the etcd
	localServices, err := getServicesFromEtcd()
	if err != nil {
		log.Slogger.Fatalf("Agent init failed =>getServicesFromEtcd Err:[%s]", err)
	} else {
		// Set services in agent
		CurAgent.SetServicesStruct(localServices)
	}

}

// getServicesFromEtcd gets the service that belongs to the current agent from the etcd
func getServicesFromEtcd() ([]executor.Service, error) {
	var servicelist []executor.Service
	servicesSlice, err := common.EtcdClient.GetWithPrefix(CurAgent.ServicePrefix)
	if err != nil {
		return nil, err
	}
	for _, ser := range servicesSlice {
		serviceStruct, err := executor.NewServiceFromJSON(ser)
		if err != nil {
			return nil, err
		}
		servicelist = append(servicelist, serviceStruct)
	}
	return servicelist, nil
}
