package redconf

import (
	"fmt"
	"strings"
	"sync"

	"github.com/garyburd/redigo/redis"
)

var (
	_ Monitor = (*RedisMonitor)(nil)
)

const (
	DefaultSubscribeChannel = "REDCONF:ONCHANGED"
)

type RedisMonitor struct {
	address  string
	password string
	channel  string
	pool     *redis.Pool

	watchingNamespace map[string]bool
	watchLocker       sync.Mutex
}

func init() {
	RegisterMonitor("redis", NewRedisMonitor)
}

func NewRedisMonitor(opts Options) (monitor Monitor, err error) {

	address := ""
	password := ""
	channel := ""

	opts.Get("address", &address)
	opts.Get("password", &password)

	if exist := opts.Get("channel", &channel); !exist {
		channel = DefaultSubscribeChannel
	}

	if address == "" {
		address = "localhost:6379"
	}

	m := &RedisMonitor{
		address:           address,
		password:          password,
		channel:           channel,
		watchingNamespace: make(map[string]bool),
	}

	m.pool = redis.NewPool(m.getRedisConn, 0)

	monitor = m

	return
}

func (p *RedisMonitor) Watch(namespace string, callback KeyContentChangedCallback, onError OnWatchingError) (err error) {

	p.watchLocker.Lock()
	defer p.watchLocker.Unlock()

	var exist bool

	if _, exist = p.watchingNamespace[namespace]; exist {
		err = fmt.Errorf("redconf: namespace of %s already in watching", namespace)
		return
	}

	go p.watchNamespace(namespace, callback, onError)

	return
}

func (p *RedisMonitor) watchNamespace(namespace string, callback KeyContentChangedCallback, onError OnWatchingError) {

	var err error

	defer func() {
		p.watchLocker.Lock()
		delete(p.watchingNamespace, namespace)
		p.watchLocker.Unlock()

		if err != nil && onError != nil {
			go onError(namespace, err)
		}
	}()

	var conn redis.Conn
	if conn, err = p.getRedisConn(); err != nil {
		return
	}

	sub := &redis.PubSubConn{conn}

	if err = sub.Subscribe(p.channel); err != nil {
		return
	}

	for {
		switch v := sub.Receive().(type) {
		case redis.Message:
			if callback != nil {
				key := string(v.Data)
				if strings.HasPrefix(key, namespace) {
					if namespace != "" {
						key = strings.TrimPrefix(key, namespace+":")
					}
					if key != "" {
						go callback(namespace, key)
					}
				}
			}
		case error:
			err = v
			return
		}
	}

	return
}

func (p *RedisMonitor) getRedisConn() (conn redis.Conn, e error) {

	conn, e = redis.Dial("tcp", p.address)
	if e != nil {
		return
	}

	if p.password != "" {
		if _, e = conn.Do("AUTH", p.password); e != nil {
			conn.Close()
			return
		}
	}

	return
}
