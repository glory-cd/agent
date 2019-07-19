package etcd

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"sync"
	"time"
	"unsafe"
	"utils/log"
)

type ServiceRegister struct {
	Client        *clientv3.Client
	Lease         clientv3.Lease
	LeaseResp     *clientv3.LeaseGrantResponse
	canclefunc    func()
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
}

func NewServiceRegister(endpoint []string, dialtimeout time.Duration) (*ServiceRegister, error) {
	conf := clientv3.Config{
		Endpoints:   endpoint,
		DialTimeout: dialtimeout * time.Second,
	}

	var client *clientv3.Client

	if clientTem, err := clientv3.New(conf); err == nil {
		client = clientTem
	} else {
		return nil, err
	}

	sr := &ServiceRegister{Client: client}
	if err := sr.setLease(10); err != nil {
		return nil, err
	}

	go sr.ListenLeaseRespChan()
	return sr, nil
}

func (sr *ServiceRegister) setLease(ttl int64) error {
	lease := clientv3.NewLease(sr.Client)

	// 设置租约时间
	leaseResp, err := lease.Grant(context.TODO(), ttl)

	if err != nil {
		return err
	}

	//	设置续租
	ctx, cancelFunc := context.WithCancel(context.TODO())
	leaseRespChan, err := lease.KeepAlive(ctx, leaseResp.ID)
	if err != nil {
		return err
	}

	sr.Lease = lease
	sr.LeaseResp = leaseResp
	sr.canclefunc = cancelFunc
	sr.keepAliveChan = leaseRespChan
	return nil

}

func (sr *ServiceRegister) ListenLeaseRespChan() {
	for {
		select {
		case leaseKeepResp := <-sr.keepAliveChan:
			if leaseKeepResp == nil {
				log.Slogger.Error("closed keepalive.")
				return
			} else {
				//log.Slogger.Debug("keepalive successful.")
			}
		}
	}
}

//注册服务
func (sr *ServiceRegister) PutService(key, val string) error {
	kv := clientv3.NewKV(sr.Client)
	_, err := kv.Put(context.TODO(), key, val, clientv3.WithLease(sr.LeaseResp.ID))
	return err
}

//撤销租约
func (sr *ServiceRegister) RevokeLease() error {
	sr.canclefunc()
	time.Sleep(2 * time.Second)
	_, err := sr.Lease.Revoke(context.TODO(), sr.LeaseResp.ID)
	return err
}

type AgentFunc func(string, string)

type ClientDis struct {
	client     *clientv3.Client
	serverList map[string]string
	lock       sync.Mutex
	putagent   AgentFunc
	delagent   AgentFunc
}

func NewClientDis(endpoint []string, dialtimeout time.Duration, putagentfunc, delagentfunc AgentFunc) (*ClientDis, error) {
	conf := clientv3.Config{
		Endpoints:   endpoint,
		DialTimeout: dialtimeout * time.Second,
	}
	if client, err := clientv3.New(conf); err == nil {
		return &ClientDis{
			client:     client,
			serverList: make(map[string]string),
			putagent:   putagentfunc,
			delagent:   delagentfunc,
		}, nil
	} else {
		return nil, err
	}
}

func (sd *ClientDis) GetService(prefix string) ([]string, error) {
	resp, err := sd.client.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	addrs := sd.extractAddrs(resp)

	go sd.watcher(prefix)
	return addrs, nil
}

func (sd *ClientDis) watcher(prefix string) {
	rch := sd.client.Watch(context.Background(), prefix, clientv3.WithPrefix())

	for wresp := range rch {
		for _, ev := range wresp.Events {
			//fmt.Println(ev.Type)
			switch ev.Type {
			case mvccpb.PUT:
				sd.SetServiceList(string(ev.Kv.Key), string(ev.Kv.Value))
			case mvccpb.DELETE:
				sd.DelServiceList(string(ev.Kv.Key), string(ev.Kv.Value))
			}
		}
	}
}

func (sd *ClientDis) extractAddrs(resp *clientv3.GetResponse) []string {
	addrs := make([]string, 0)
	if resp == nil || resp.Kvs == nil {
		return addrs
	}
	for i := range resp.Kvs {
		if v := resp.Kvs[i].Value; v != nil {
			key := resp.Kvs[i].Key
			sd.SetServiceList(string(key), string(v))
			addrs = append(addrs, string(key))
		}
	}
	return addrs
}

func (this *ClientDis) SetServiceList(key, val string) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.serverList[key] = string(val)
	log.Slogger.Debug("set data key :", key, "val:", val)
	go this.putagent(key, val)
}

func (this *ClientDis) DelServiceList(key, val string) {
	this.lock.Lock()
	defer this.lock.Unlock()
	delete(this.serverList, key)
	log.Slogger.Debug("del data key:", key)
	go this.delagent(key, val)
}

type AfisServiceRegister struct {
	Client *clientv3.Client
}

func (asr *AfisServiceRegister) NewAfisServiceRegister(endpoint []string, dialtimeout time.Duration) (*AfisServiceRegister, error) {
	conf := clientv3.Config{
		Endpoints:   endpoint,
		DialTimeout: dialtimeout * time.Second,
	}

	var client *clientv3.Client

	if clientTem, err := clientv3.New(conf); err == nil {
		client = clientTem
	} else {
		return nil, err
	}

	sr := &AfisServiceRegister{Client: client}

	return sr, nil
}

func (asr *AfisServiceRegister) GetWithPrefix(prefix string) ([]string, error) {
	var re []string
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	resp, err := asr.Client.Get(ctx, prefix, clientv3.WithPrefix())
	cancel()
	if err != nil {
		return nil, err
	}
	for _, ev := range resp.Kvs {
		re = append(re, *(*string)(unsafe.Pointer(&ev.Value)))
	}

	return re, nil
}

func (asr *AfisServiceRegister) Put(key, value string) error {
	//设置1秒超时，访问etcd有超时控制
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//操作etcd
	_, err := asr.Client.Put(ctx, key, value)
	//操作完毕，取消etcd
	cancel()
	if err != nil {
		return err
	}
	return nil
}

/*
	如果isprefix是true,则根据前缀获取val值，反之isprefix是false,则根据key值获取val值
    返回返回map[key]val,error
*/
func (asr *AfisServiceRegister) Get(key string, isprefix bool) (map[string]string, error) {
	//设置1秒超时，访问etcd有超时控制
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var resp *clientv3.GetResponse
	var err error

	if isprefix {
		resp, err = asr.Client.Get(ctx, key, clientv3.WithPrefix())
	} else {
		resp, err = asr.Client.Get(ctx, key)
	}

	if err != nil {
		return nil, err
	}
	kvmap := make(map[string]string, 0)
	if resp != nil && resp.Kvs != nil {
		for i := range resp.Kvs {
			if v := resp.Kvs[i].Value; v != nil {
				key := resp.Kvs[i].Key
				kvmap[string(key)] = string(v)
			}
		}
	}
	return kvmap, nil
}