package executor

import (
	"testing"
)

func TestRollBackNormal(t *testing.T) {
	u := &Upgrade{Service: &Service{Dir: "/home/alarm/AlarmWechat", OsUser:"alarm"}, backfile: "/home/alarm/Backup/AlarmWechat20190524110559.143.zip"}
	err := u.rollBack()

	if err != nil {
		t.Errorf(`Download("/home/alarm/AlarmWechat %s") = error`, err.Error())
	}
}
