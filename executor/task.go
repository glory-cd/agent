package executor

import "github.com/auto-cdp/agent/common"

func (t *Executor) Execute() (resultJson string) {
	var dr Drive
	result := NewResultPointer(t.Identiy)
	switch t.OP {
	case common.Operate_DEP:
		dr = &Deploy{driver: driver{Task: t.Task, Service: t.Service}, rs: *result}
	case common.Operate_UPG:
		dr = &Upgrade{driver: driver{Task: t.Task, Service: t.Service}, rs: *result}
	case common.Operate_STA, common.Operate_SHU, common.Operate_RES:
		dr = &Rss{driver: driver{Task: t.Task, Service: t.Service}, rs: *result}
	case common.Operate_CHE:
		dr = &Check{driver: driver{Task: t.Task, Service: t.Service}, rs: *result}
	case common.Operate_BAK:
		dr = &Backup{driver: driver{Task: t.Task, Service: t.Service}, rs: *result}
	default:
		return
	}

	ch := make(chan Result)

	go dr.Exec(ch)

	re := <-ch

	restring, _ := re.ToJsonString()

	return restring
}
