/**
* @Author: xhzhang
* @Date: 2019-04-18 17:22
 */
package main

import (
	"github.com/glory-cd/agent/core"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Set the command line Flag
var (
	etcdAddress = kingpin.Flag("etcd", "ETCD address to connect.").Short('e').Required().String()
)

func main() {
	kingpin.Version("Version: 0.0.13")
	kingpin.Parse()
	// Initialize
	core.InitAgent(*etcdAddress)
	// Run
	core.Run()
}
