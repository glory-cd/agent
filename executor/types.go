package executor

import "agent/common"

type Drive interface {
	Exec(out chan<- Result)
}

type Identiy struct {
	TaskID      int `json:"taskid"`
	ExecutionID int `json:"executionid"`
}

type Task struct {
	Identiy
	OP            common.OpMode `json:"serviceop"`
	CustomPattern []string      `json:"servicecustompattern"`
	RemoteCode    string        `json:"serviceremotecode"`
}

type Service struct {
	ServiceID   string   `json:"serviceid"`
	Name        string   `json:"servicename"`
	OsUser      string   `json:"serviceosuser"`
	Dir         string   `json:"servicedir"`
	ModuleName  string   `json:"servicemodulename"`
	CodePattern []string `json:"servicecodepattern"`
	PidFile     string   `json:"servicepidfile"`
	StartCMD    string   `json:"servicestartcmd"`
	StopCMD     string   `json:"servicestopcmd"`
}

type Executor struct {
	*Task
	*Service
}
