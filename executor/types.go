package executor

import "github.com/glory-cd/agent/common"

// Driver interface
type Drive interface {
	Exec(out chan<- Result)
}

// Task identity
type Identiy struct {
	TaskID      int `json:"taskid"`
	ExecutionID int `json:"executionid"`
}

// Task
type Task struct {
	Identiy
	OP            common.OpMode `json:"serviceop"`
	CustomPattern []string      `json:"servicecustompattern"`
	RemoteCode    string        `json:"serviceremotecode"`
	UserPass      string        `json:"serviceosuserpass"`
}

// Service
type Service struct {
	ServiceID   string   `json:"serviceid"`
	OsUser      string   `json:"serviceosuser"`
	Dir         string   `json:"servicedir"`
	ModuleName  string   `json:"servicemodulename"`
	CodePattern []string `json:"servicecodepattern"`
	PidFile     string   `json:"servicepidfile"`
	StartCMD    string   `json:"servicestartcmd"`
	StopCMD     string   `json:"servicestopcmd"`
}

// Executor
type Executor struct {
	*Task
	*Service
}

type User struct{
	// Uid is the user ID.
	Uid int
	// Gid is the primary group ID.
	Gid int
	// Username is the login name.
	Username string
	// Name is the user's real or display name.
	Name string
	// HomeDir is the path to the user's home directory (if they have one).
	HomeDir string
}