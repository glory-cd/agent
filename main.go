/**
* @Author: xhzhang
* @Date: 2019-04-18 17:22
 */
package main

import (
	"github.com/auto-cdp/agent/core"
	"gopkg.in/alecthomas/kingpin.v2"
)

//设置命令行Flag
var (
	etcdAddress = kingpin.Flag("etcd", "ETCD address to connect.").Short('e').Required().String()
)

func main() {
	kingpin.Version("Version: 0.0.1")
	kingpin.Parse()
	//初始化服务
	core.InitAgent(*etcdAddress)
	//服务运行
	core.Run()
}
