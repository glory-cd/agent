// Package listener handles instruction
package listener

import (
	"encoding/json"
	"fmt"
	"github.com/glory-cd/agent/common"
	"github.com/glory-cd/agent/executor"
	"github.com/glory-cd/agent/executor/backup"
	"github.com/glory-cd/agent/executor/check"
	"github.com/glory-cd/agent/executor/delete"
	"github.com/glory-cd/agent/executor/deploy"
	"github.com/glory-cd/agent/executor/roll"
	"github.com/glory-cd/agent/executor/rss"
	"github.com/glory-cd/agent/executor/upgrade"
	"github.com/glory-cd/utils/log"
	"strconv"
	"strings"
)

// Executor interface
type Executor interface {
	Exec(rs *executor.Result)
}

// According to different instructions, call the corresponding driver
func execute(ex Executor) (resultJSON string) {
	result := executor.NewResult()
	// execute
	ex.Exec(result)
	// convert to json
	resultJSON, _ = result.ToJSONString()
	return
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

// dealReceiveInstruction handles received instructions
func dealReceiveInstruction(ins string) {
	var insDriver executor.Driver
	err := json.Unmarshal([]byte(ins), &insDriver)
	if err != nil {
		log.Slogger.Errorf("ConvertInsJsonTOTaskObject Err:[%s]", err.Error())
		return
	}
	log.Slogger.Debugf("Recived Instruction task: %+v, service: %+v",
		*insDriver.Task, *insDriver.Service)

	dynamicExecutor := getExecutor(insDriver)

	if dynamicExecutor == nil {
		log.Slogger.Error("Expect to get a concrete executor, but return nil!")
		return
	}
	//execute
	result := execute(dynamicExecutor)

	publishResult(insDriver.TaskID, result)
}

// getExecutor instantiates an concrete executor
func getExecutor(insDriver executor.Driver) Executor {

	var dynamicExecutor Executor
	switch insDriver.OP {
	case common.OperateDEP:
		dynamicExecutor = deploy.NewDeploy(insDriver)
	case common.OperateUPG:
		dynamicExecutor = upgrade.NewUpgrade(insDriver)
	case common.OperateSTA, common.OperateSHU, common.OperateRES:
		dynamicExecutor = rss.NewRss(insDriver)
	case common.OperateCHE:
		dynamicExecutor = check.NewCheck(insDriver)
	case common.OperateBAK:
		dynamicExecutor = backup.NewBackup(insDriver)
	case common.OperateROL:
		dynamicExecutor = roll.NewRoll(insDriver)
	case common.OperateDEL:
		dynamicExecutor = delete.NewDelete(insDriver)
	default:
		return nil
	}

	return dynamicExecutor
}

func annulService() error {

	return nil
}
