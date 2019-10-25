/**
* @Author: xhzhang
* @Date: 2019-04-19 10:25
 */
package core

import (
	"encoding/json"
	"fmt"
	"github.com/glory-cd/agent/common"
	"github.com/glory-cd/agent/executor"
	"github.com/glory-cd/utils/log"
	"strconv"
	"strings"
)

//Process received instructions
func dealReceiveInstruction(ins string) {
	var insExecutor executor.Executor
	err := json.Unmarshal([]byte(ins), &insExecutor)
	if err != nil {
		log.Slogger.Errorf("ConvertInsJsonTOTaskObject Err:[%s]", err.Error())
		return
	}
	log.Slogger.Debugf("Recived Instruction task: %+v, service: %+v", *insExecutor.Task, *insExecutor.Service)
	//Execute
	result := insExecutor.Execute()

	publishResult(insExecutor.TaskID, result)
}

//Push the results to redis
func publishResult(taskid int, re string) {

	log.Slogger.Infof("push result to redis : %s", re)
	resultChanel := strings.Join([]string{"result", strconv.Itoa(taskid)}, ".")
	fmt.Printf("chanel:%s\n", resultChanel)
	num, err := common.RedisConn.Publish(resultChanel, re)
	if err != nil {
		log.Slogger.Error(err.Error(), num)
	}
}
