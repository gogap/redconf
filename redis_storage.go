package redconf

import (
	"github.com/garyburd/redigo/redis"
)

var (
	_ Storage = (*RedisStorage)(nil)
)

type RedisStorage struct {
	address  string
	password string
	db       int
	idle     int
	pool     *redis.Pool
}

func init() {
	RegisterStorage("redis", NewRedisStorage)
}

func NewRedisStorage(opts Options) (storage Storage, err error) {

	address := ""
	password := ""
	db := 0
	idle := 0

	opts.Get("address", &address)
	opts.Get("password", &password)
	opts.Get("db", &db)
	opts.Get("idle", &idle)

	if address == "" {
		address = "localhost:6379"
	}

	if idle == 0 {
		idle = 5
	}

	s := &RedisStorage{
		address:  address,
		password: password,
		db:       db,
		idle:     idle,
	}

	s.pool = redis.NewPool(s.getRedisConn, 0)

	storage = s

	return
}

func (p *RedisStorage) getRedisKey(namespace, key string) string {
	if namespace == "" {
		return key
	}

	return namespace + ":" + key
}

func (p *RedisStorage) Set(namespace, key string, val interface{}) (err error) {

	conn := p.pool.Get()

	if _, err = conn.Do("SET", p.getRedisKey(namespace, key), val); err != nil {
		return
	}

	return
}

func (p *RedisStorage) Get(namespace, key string) (ret interface{}, err error) {

	conn := p.pool.Get()

	var reply interface{}

	if reply, err = conn.Do("GET", p.getRedisKey(namespace, key)); err != nil {
		return
	}

	if ret, err = redis.String(reply, err); err != nil {
		if err == redis.ErrNil {
			err = nil
		}
	}

	return
}

func (p *RedisStorage) getRedisConn() (conn redis.Conn, e error) {

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

	if _, e = conn.Do("SELECT", p.db); e != nil {
		return
	}

	return
}
