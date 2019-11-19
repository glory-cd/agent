package executor

import (
	"bytes"
	"encoding/json"
	"github.com/glory-cd/agent/common"
	"github.com/glory-cd/utils/afis"
)


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

// Build a Service with ServiceID from json string
func NewServiceFromJson(sjson string) (Service, error) {
	var s Service
	err := json.Unmarshal([]byte(sjson), &s)
	if err != nil {
		return s, err
	}
	s.ServiceID = afis.GetMd5String(common.AgentID + s.Dir)
	return s, nil
}

// convert Service to json string
func NewJsonFromService(s Service) (string, error) {
	jsonb, err := json.Marshal(s)
	sjson := bytes.NewBuffer(jsonb).String()
	if err != nil {
		return "", err
	}
	return sjson, nil
}