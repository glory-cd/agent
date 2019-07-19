/**
* @Author: xhzhang
* @Date: 2019-04-19 10:25
 */
package core

import (
	"encoding/json"
	"fmt"
	"github.com/auto-cdp/agent/common"
	"github.com/auto-cdp/agent/executor"
	"github.com/auto-cdp/utils/log"
	"strconv"
	"strings"
)

func dealReceiveInstruction(ins string) {
	var insExecutor executor.Executor
	err := json.Unmarshal([]byte(ins), &insExecutor)
	if err != nil {
		log.Slogger.Errorf("ConvertInsJsonTOTaskObject Err:[%s]", err.Error())
		return
	}
	result := insExecutor.Execute()

	publishResult(insExecutor.TaskID, result)
}

//向redis推送结果
func publishResult(taskid int, re string) {

	log.Slogger.Infof("push result to redis : %s", re)
	resultChanel := strings.Join([]string{"result", strconv.Itoa(taskid)}, ".")
	fmt.Printf("chanel:%s\n", resultChanel)
	num, err := common.RedisConn.Publish(resultChanel, re)
	if err != nil {
		log.Slogger.Error(err.Error(), num)
	}
}
