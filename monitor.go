package redconf

import (
	"errors"
	"sync"
)

type KeyContentChangedCallback func(namespace, key string)

type OnWatchingError func(namespace string, err error)

type NewMonitorFunc func(opts Options) (monitor Monitor, err error)

type Monitor interface {
	Watch(namespace string, callback KeyContentChangedCallback, onError OnWatchingError) (err error)
}

var (
	monitorDrivers map[string]NewMonitorFunc = make(map[string]NewMonitorFunc)

	monitorLocker sync.Mutex
)

func RegisterMonitor(dirverName string, newFunc NewMonitorFunc) (err error) {
	if dirverName == "" {
		err = errors.New("redconf: driver name could not be empty")
		return
	}

	if newFunc == nil {
		err = errors.New("redconf: new monitor func could not be nil")
		return
	}

	monitorLocker.Lock()
	defer monitorLocker.Unlock()

	if _, exist := monitorDrivers[dirverName]; exist {
		err = errors.New("redconf: monitor driver of " + dirverName + " already exist")
		return
	}

	monitorDrivers[dirverName] = newFunc

	return
}

func CreateMonitor(driverName string, opts Options) (monitor Monitor, err error) {
	if driverName == "" {
		err = errors.New("redconf: driver name could not be empty")
		return
	}

	var exist bool
	var newFunc NewMonitorFunc

	if newFunc, exist = monitorDrivers[driverName]; !exist {
		err = errors.New("redconf: monitor driver of " + driverName + " not exist")
		return
	}

	if monitor, err = newFunc(opts); err != nil {
		return
	}

	return
}
