package listener

import (
	"github.com/glory-cd/agent/common"
	"github.com/glory-cd/utils/etcd"
	"github.com/glory-cd/utils/log"
)

func Run(){
	// Register the agent to etcd
	go startRegister()

	// Subscribe redis channel to receive instructions
	go subscribeCMDChannel()

	// Start the restful goroutine to process registration behavior
	go startRestful()

	// Start listening for Grace signals
	gracefulHandle()
}

func startRegister() {
	agentregister, err := etcd.NewServiceRegister(common.EtcdEndpoint, common.DialTimeout)
	if err != nil {
		log.Slogger.Fatalf("agent resister fail.[%s]", err)
	}

	err = agentregister.PutService(CurAgent.EtcdKey, CurAgent.EtcdVal)
	if err != nil {
		log.Slogger.Fatalf("agent register fail.[%s]", err)
	}

	log.Slogger.Infof("agent register successful! [%s] : [%s]", CurAgent.EtcdKey, CurAgent.EtcdVal)

}

// Subscribe signal channel
// and instruction channel
func subscribeCMDChannel() {
	go common.RedisConn.SubscribeChannel(CurAgent.InstructionChannel, dealReceiveInstruction)
	go common.RedisConn.SubscribeChannel(CurAgent.GraceChannel, dealReceiveGraceCMD)
}
