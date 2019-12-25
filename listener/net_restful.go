package listener

import (
	"bytes"
	"github.com/glory-cd/agent/common"
	"github.com/glory-cd/agent/executor"
	"github.com/glory-cd/utils/log"
	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

func startRestful() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/register", DealRecieveService)
	log.Slogger.Fatal(http.ListenAndServe(common.Config().Rest.Addr, router))
}

// Register handler
func DealRecieveService(w http.ResponseWriter, r *http.Request) {

	var err error
	defer func() {
		if err != nil{
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}else{
			http.Error(w, "Register successful", http.StatusOK)
		}
	}()
	// Process incoming service information
	result, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Slogger.Warn("ReadRequestBody Err:[%s]", err)
		return
	}
	sjson := bytes.NewBuffer(result).String()
	log.Slogger.Debugf("recive service from script: %s", sjson)

	service, err := executor.NewServiceFromJson(sjson)
	if err != nil {
		log.Slogger.Warn("ServiceJsonToServiceStruct Err:[%s]", err)
		return
	}

	// If it is a new service, register directly to etcd
	if !CurAgent.CheckRegisterIsExist(service.ServiceID) {
		err = writeJson(service)
		if err != nil {
			log.Slogger.Error("Etcd PUT Falied Err:[%s]", err)
			return
		}
		// Add to the Agent
		CurAgent.AddService(service)
		log.Slogger.Debugf("New Register successful: %s", service.ServiceID)
		return
	}

	// If anything changes, sync to etcd
	if !cmp.Equal(service, CurAgent.GetService(service.ServiceID)) {
		err = writeJson(service)
		if err != nil {
			log.Slogger.Error("Etcd PUT Falied Err:[%s]", err)
			return
		}
		//Sync to Agent
		CurAgent.SyncService(service)
		log.Slogger.Debugf("Sync Register successful: %s", service.ServiceID)
		return
	}

	log.Slogger.Debugf("Do not need to Register :: %s", service.ServiceID)
	return
}

// writeJson Puts Executor.Service  to etcd in json format
func writeJson(s executor.Service) error {
	jsonWithId, err := executor.NewJsonFromService(s)
	if err != nil {
		return err
	}

	key := CurAgent.ServicePrefix + s.ServiceID
	err = common.EtcdClient.Put(key, jsonWithId)
	if err != nil {
		return err
	}
	return nil
}
