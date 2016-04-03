package redconf

import (
	"fmt"
	"sync"
	"time"
)

type RedConf struct {
	namespace string

	watching map[string]*WatchingConfig

	watchingKeyIndex map[string]*Field

	storage Storage
	monitor Monitor

	confLock sync.Mutex
}

func New(namespace string, storage Storage, monitor Monitor) (redConf *RedConf, err error) {

	if storage == nil {
		err = fmt.Errorf("redconf: the storage is nil")
		return
	}

	if monitor == nil {
		err = fmt.Errorf("redconf the monitor is nil")
		return
	}

	redConf = &RedConf{
		namespace:        namespace,
		storage:          storage,
		monitor:          monitor,
		watching:         make(map[string]*WatchingConfig),
		watchingKeyIndex: make(map[string]*Field),
	}

	return
}

func (p *RedConf) Watch(vals ...interface{}) (err error) {

	if vals == nil {
		return
	}

	var confs []*WatchingConfig
	for _, val := range vals {
		var wConf *WatchingConfig
		if wConf, err = NewWatchingConfig(val); err != nil {
			return
		}
		confs = append(confs, wConf)
	}

	return p.WatchWithConfig(confs...)
}

func (p *RedConf) Keys() []string {
	var keys []string

	for k, _ := range p.watchingKeyIndex {
		keys = append(keys, k)
	}

	return keys
}

func (p *RedConf) WatchWithConfig(configs ...*WatchingConfig) (err error) {

	p.confLock.Lock()
	defer p.confLock.Unlock()

	var watchingKeys []string

	for _, conf := range configs {

		if wConf, exist := p.watching[conf.name]; exist {
			if wConf.value != conf.value {
				err = fmt.Errorf("redconf: watch config of %s already exist", wConf.name)
				return
			}
		}
		p.watching[conf.name] = conf

		for _, field := range conf.fields {
			watchingKeys = append(watchingKeys, field.String())

			var oldF *Field
			var exist bool
			if oldF, exist = p.watchingKeyIndex[field.String()]; !exist {
				p.watchingKeyIndex[field.String()] = field
			} else if oldF != field {
				err = fmt.Errorf("redconf: the key of %s in namespace %s already have struct to watch", field.String(), p.namespace)
				return
			}
		}
	}

	if len(watchingKeys) > 0 {

		if err = p.syncKeys(watchingKeys...); err != nil {
			return
		}

		if err = p.monitor.Watch(p.namespace, p.onKeyContentChanged, p.onMonitorError); err != nil {
			return
		}
	}

	return
}

func (p *RedConf) Namespace() string {
	return p.namespace
}

func (p *RedConf) syncKeys(keys ...string) (err error) {

	var kvs map[string]interface{} = make(map[string]interface{})

	for _, key := range keys {
		var val interface{}
		if val, err = p.storage.Get(p.namespace, key); err != nil {
			return
		}
		kvs[key] = val
	}

	for k, v := range kvs {
		if err = p.setFieldValue(k, v); err != nil {
			return
		}
	}

	return
}

func (p *RedConf) onKeyContentChanged(namespace, key string) {
	if namespace != p.namespace {
		return
	}

	var err error
	var value interface{}

	if value, err = p.storage.Get(namespace, key); err != nil {
		return
	}

	p.setFieldValue(key, value)
}

func (p *RedConf) onMonitorError(namespace string, err error) {
	if namespace != p.namespace {
		return
	}

	time.Sleep(time.Second * 5)

	p.monitor.Watch(p.namespace, p.onKeyContentChanged, p.onMonitorError)
}

func (p *RedConf) setFieldValue(keyName string, value interface{}) (err error) {
	var field *Field
	var exist bool
	if field, exist = p.watchingKeyIndex[keyName]; !exist {
		return
	}

	var fv interface{}
	if fv, err = conv(field.Type(), value); err != nil {
		return
	}

	field.set(fv)

	return
}
