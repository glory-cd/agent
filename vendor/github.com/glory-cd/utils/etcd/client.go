/**
* @Author: xhzhang
* @Date: 2019-06-11 14:21
 */
package etcd

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"github.com/glory-cd/utils/log"
	"time"
)

type DealFunc func(string, string, string)

type BaseClient struct {
	client *clientv3.Client
}

func NewBaseClient(endpoint []string, dialtimeout time.Duration) (*BaseClient, error) {
	conf := clientv3.Config{
		Endpoints:   endpoint,
		DialTimeout: dialtimeout * time.Second,
	}
	if client, err := clientv3.New(conf); err == nil {
		return &BaseClient{
			client: client,
		}, nil
	} else {
		return nil, err
	}
}

/*
	写入key,val
	返回错误信息
*/

func (c *BaseClient) Put(key, value string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := c.client.Put(ctx, key, value)
	return err
}

/*
	如果isprefix是true,则根据前缀获取val值，反之isprefix是false,则根据key值获取val值
    返回返回map[key]val,error
*/
func (c *BaseClient) Get(key string, isprefix bool) (map[string]string, error) {
	//设置1秒超时，访问etcd有超时控制
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var resp *clientv3.GetResponse
	var err error

	if isprefix {
		resp, err = c.client.Get(ctx, key, clientv3.WithPrefix())
	} else {
		resp, err = c.client.Get(ctx, key)
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

func (c *BaseClient) Del(info string, isprefix bool) error {
	//设置1秒超时，访问etcd有超时控制
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var err error
	if isprefix {
		_, err = c.client.Delete(ctx, info, clientv3.WithPrefix())
	} else {
		_, err = c.client.Delete(ctx, info)
	}
	return err
}

func (bc *BaseClient) GetAgents(dealfunc DealFunc) (map[string]string, error) {
	prefix := "/agent/"
	agentmap, err := bc.Get(prefix, true)
	if err != nil {
		return nil, err
	}

	go bc.watcher(prefix, true, dealfunc)
	return agentmap, nil
}

func (bc *BaseClient) GetServices(dealfunc DealFunc) (map[string]string, error) {
	prefix := "/service/"
	servicemap, err := bc.Get(prefix, true)
	if err != nil {
		return nil, err
	}

	go bc.watcher(prefix, true, dealfunc)
	return servicemap, nil
}

func (c *BaseClient) watcher(key string, isprefix bool, deal DealFunc) {
	var rch clientv3.WatchChan
	if isprefix {
		rch = c.client.Watch(context.Background(), key, clientv3.WithPrefix())
	} else {
		rch = c.client.Watch(context.Background(), key)
	}
	for wresp := range rch {
		for _, ev := range wresp.Events {
			deal(ev.Type.String(), string(ev.Kv.Key), string(ev.Kv.Value))
		}
	}
}

type AfisRegister struct {
	BClient       BaseClient
	Lease         clientv3.Lease
	LeaseResp     *clientv3.LeaseGrantResponse
	canclefunc    func()
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
}

func NewAfisRegister(bc BaseClient) (*AfisRegister, error) {
	ar := &AfisRegister{BClient: bc}
	if err := ar.setLease(10); err != nil {
		return nil, err
	}

	go ar.ListenLeaseRespChan()
	return ar, nil
}

func (ar *AfisRegister) setLease(ttl int64) error {
	lease := clientv3.NewLease(ar.BClient.client)

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

	ar.Lease = lease
	ar.LeaseResp = leaseResp
	ar.canclefunc = cancelFunc
	ar.keepAliveChan = leaseRespChan
	return nil

}

//注册
func (ar *AfisRegister) PutWithLease(key, val string) error {
	kv := clientv3.NewKV(ar.BClient.client)
	_, err := kv.Put(context.TODO(), key, val, clientv3.WithLease(ar.LeaseResp.ID))
	return err
}

func (ar *AfisRegister) ListenLeaseRespChan() {
	for {
		select {
		case leaseKeepResp := <-ar.keepAliveChan:
			if leaseKeepResp == nil {
				log.Slogger.Error("closed keepalive.")
				return
			} else {
				log.Slogger.Debug("keepalive successful.")
			}
		}
	}
}

//撤销租约
func (ar *AfisRegister) RevokeLease() error {
	ar.canclefunc()
	time.Sleep(2 * time.Second)
	_, err := ar.Lease.Revoke(context.TODO(), ar.LeaseResp.ID)
	return err
}
