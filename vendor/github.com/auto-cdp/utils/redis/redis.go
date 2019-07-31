package redis

import (
	"context"
	"fmt"
	"github.com/glory-cd/utils/log"
	"github.com/garyburd/redigo/redis"
	"time"
)

type DataFunc func(string)

func NewRedisPool(host string, MaxIdle int, MaxActive int, idletimeout time.Duration) redis.Pool {
	return redis.Pool{
		MaxIdle:     MaxIdle,
		MaxActive:   MaxActive,
		IdleTimeout: idletimeout * time.Second,
		Dial: func() (conn redis.Conn, e error) {
			c, err := redis.Dial("tcp", host)
			if err != nil {
				return nil, fmt.Errorf("redis connection error: %s", err)
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			if err != nil {
				return fmt.Errorf("redis connection error: %s", err)
			}
			return nil
		},
	}
}

type RedisConn struct {
	pool redis.Pool
}

func NewRedisConn(pool redis.Pool) RedisConn {
	return RedisConn{pool: pool}
}

func (rp RedisConn) Publish(channel, message string) (int, error) {
	conn := rp.pool.Get()
	defer conn.Close()

	n, err := redis.Int(conn.Do("PUBLISH", channel, message))

	return n, err
}

func (rp RedisConn) SubscribeChannel(channel string, handleMessage DataFunc) {
	conn := rp.pool.Get()
	psc := redis.PubSubConn{conn}
	if err := psc.Subscribe(channel); err != nil {
		log.Slogger.Errorf("redis subscribe fail.[%s]", err)
		return
	}
	log.Slogger.Debugf("redis subscribe successful.[%s]", channel)
	ctx := context.TODO()
	done := make(chan error, 1)
	go rp.subcribe(ctx, psc, handleMessage, done)
	err := healthchek(ctx, psc, done)
	if err != nil {
		log.Slogger.Error(err)
	}
}

func (rp RedisConn) subcribe(ctx context.Context, psc redis.PubSubConn, handleData DataFunc, done chan error) {
	defer psc.Close()
	for {
		switch msg := psc.Receive().(type) {
		case redis.Message:
			log.Slogger.Debugf(fmt.Sprintf("%s: message: %s", msg.Channel, msg.Data))
			log.Slogger.Info("启动goroutine处理.")
			go handleData(string(msg.Data))
		case redis.Subscription:
			if msg.Count == 0 {
				done <- nil
				return
			}
		case error:
			done <- fmt.Errorf("redis pubsub receiver err:%v", msg)
			log.Slogger.Error(msg)
			return
		}
	}

}

func healthchek(ctx context.Context, psc redis.PubSubConn, done chan error) error {
	tick := time.NewTicker(time.Minute)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			if err := psc.Unsubscribe(); err != nil {
				return fmt.Errorf("redis pubsub unsubscribe err %v", err)
			}
			return nil
		case err := <-done:
			return err
			/*case <-tick.C:
			if err := psc.Ping("ping"); err != nil {
				return nil
			}*/

		}
	}
}
