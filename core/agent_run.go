package core

import (
	"agent/common"
	"utils/etcd"
	"utils/log"
)

func Run(){
	//将agent注册到etcd
	go startRegister()

	//订阅redis通道以接收指令
	go subscribeCMDChannel()

	//启动restful处理注册行为
	go startRestful()

	//开始监听Grace信号
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

//分别订阅部署、更新、静态命令消息
func subscribeCMDChannel() {
	go common.RedisConn.SubscribeChannel(CurAgent.InstructionChannel, dealReceiveInstruction)
	go common.RedisConn.SubscribeChannel(CurAgent.GraceChannel, dealReceiveGraceCMD)
}
