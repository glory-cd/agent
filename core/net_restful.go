package core

import (
	"bytes"
	"fmt"
	"github.com/auto-cdp/agent/common"
	"github.com/auto-cdp/agent/executor"
	"github.com/auto-cdp/utils/log"
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
	// 处理发送过来的服务信息
	result, err := ioutil.ReadAll(r.Body)
	sjson := bytes.NewBuffer(result).String()
	log.Slogger.Debug("recive service from script." + sjson)

	service, err := executor.NewServiceFromJson(sjson)
	if err != nil {
		log.Slogger.Warn("ServiceJsonToServiceStruct Err:[%s]", err)
		_, _ = fmt.Fprintf(w, "ServiceJsonToServiceStruct Err:[%s]\n", err.Error())
		return
	}

	//如果是新服务则直接注册到etcd
	if !CurAgent.CheckRegisterIsExist(service.ServiceID) {
		// 服务信息注册到etcd
		err := writeJson(service)
		if err != nil {
			log.Slogger.Error("Etcd PUT Falied Err:[%s]", err)
			return
		}
		//增加到Agent中
		CurAgent.AddService(service)

		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, "New Register successful: %s\n", service.ServiceID)
		return
	}

	//如果有变化就写入etcd
	if !cmp.Equal(service, CurAgent.GetService(service.ServiceID)) {
		err := writeJson(service)
		if err != nil {
			log.Slogger.Error("Etcd PUT Falied Err:[%s]", err)
		}
		//同步到Agent
		CurAgent.SyncService(service)

		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, "Sync Register successful: %s\n", service.ServiceID)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "Do not need to Register : %s\n", service.ServiceID)
	return
}

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
