package executor

import "github.com/glory-cd/agent/common"

// According to different instructions, call the corresponding driver
func (t *Executor) Execute() (resultJson string) {
	var dr Drive
	result := NewResultPointer(t.Identiy)
	switch t.OP {
	case common.OperateDEP:
		dr = &Deploy{driver: driver{Task: t.Task, Service: t.Service}, rs: *result}
	case common.OperateUPG:
		dr = &Upgrade{driver: driver{Task: t.Task, Service: t.Service}, rs: *result}
	case common.OperateSTA, common.OperateSHU, common.OperateRES:
		dr = &Rss{driver: driver{Task: t.Task, Service: t.Service}, rs: *result}
	case common.OperateCHE:
		dr = &Check{driver: driver{Task: t.Task, Service: t.Service}, rs: *result}
	case common.OperateBAK:
		dr = &Backup{driver: driver{Task: t.Task, Service: t.Service}, rs: *result}
	case common.OperateROL:
		dr = &Roll{driver: driver{Task: t.Task, Service: t.Service}, rs: *result}
	default:
		return
	}

	// Prepare result channel
	ch := make(chan Result)
	// execute
	go dr.Exec(ch)
	// wait for the result
	re := <-ch
	// convert to json
	restring, _ := re.ToJsonString()

	return restring
}
