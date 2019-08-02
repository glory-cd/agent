package executor

import "github.com/glory-cd/agent/common"

//根据不同的指令，调用相应的驱动
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
	case common.Operate_ROL:
		dr = &Roll{driver: driver{Task: t.Task, Service: t.Service}, rs: *result}
	default:
		return
	}
	//准备结果通道
	ch := make(chan Result)
	//驱动执行
	go dr.Exec(ch)
	//等待接受结果
	re := <-ch
	//转为JSON
	restring, _ := re.ToJsonString()

	return restring
}
