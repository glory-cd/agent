package executor

import "agent/common"

func (t *Executor) Execute() (resultJson string) {
	var dr Drive
	result := NewResultPointer(t.Identiy)
	switch t.OP {
	case common.Operate_DEP:
		dr = &Deploy{Task: t.Task, Service: t.Service, rs: *result}
	case common.Operate_UPG:
		dr = &Upgrade{Task: t.Task, Service: t.Service, rs: *result}
	case common.Operate_STA, common.Operate_SHU, common.Operate_RES:
		dr = &Rss{Task: t.Task, Service: t.Service, rs: *result}
	case common.Operate_CHE:
		dr = &Check{Task: t.Task, Service: t.Service, rs: *result}
	default:
		return
	}

	ch := make(chan Result)

	go dr.Exec(ch)

	re := <-ch

	restring, _ := re.ToJsonString()

	return restring
}
