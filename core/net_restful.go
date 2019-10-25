package core

import (
	"bytes"
	"fmt"
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

func DealRecieveService(w http.ResponseWriter, r *http.Request) {

	var err error
	defer func() {
		if err != nil{

		}else{

		}
	}()
	// Process incoming service information
	result, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Slogger.Warn("ReadRequestBody Err:[%s]", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, "ReadRequestBody Err:[%s]\n", err.Error())
		return
	}
	sjson := bytes.NewBuffer(result).String()
	log.Slogger.Debugf("recive service from script: %s", sjson)

	service, err := executor.NewServiceFromJson(sjson)
	if err != nil {
		log.Slogger.Warn("ServiceJsonToServiceStruct Err:[%s]", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, "ServiceJsonToServiceStruct Err:[%s]\n", err.Error())
		return
	}

	//如果是新服务则直接注册到etcd
	if !CurAgent.CheckRegisterIsExist(service.ServiceID) {
		// 服务信息注册到etcd
		err = writeJson(service)
		if err != nil {
			log.Slogger.Error("Etcd PUT Falied Err:[%s]", err)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprintf(w, "Etcd PUT Falied Err:[%s]\n", err.Error())
			return
		}
		//增加到Agent中
		CurAgent.AddService(service)
		log.Slogger.Debugf("New Register successful: %s", service.ServiceID)
		//设置响应头，并响应请求
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, "New Register successful: %s\n", service.ServiceID)
		return
	}

	//如果有变化就写入etcd
	if !cmp.Equal(service, CurAgent.GetService(service.ServiceID)) {
		err = writeJson(service)
		if err != nil {
			log.Slogger.Error("Etcd PUT Falied Err:[%s]", err)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprintf(w, "Etcd PUT Falied Err:[%s]\n", err)
			return
		}
		//同步到Agent
		CurAgent.SyncService(service)
		log.Slogger.Debugf("Sync Register successful: %s", service.ServiceID)
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, "Sync Register successful: %s\n", service.ServiceID)

		return
	}

	log.Slogger.Debugf("Do not need to Register :: %s", service.ServiceID)
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "Do not need to Register : %s\n", service.ServiceID)
	return
}

//将executor.Service以json格式写入etcd
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
