package redconf

import (
	"errors"
	"sync"
)

type NewStorageFunc func(opts Options) (storage Storage, err error)

type Storage interface {
	Set(namespace, key string, val interface{}) (err error)
	Get(namespace, key string) (ret interface{}, err error)
}

var (
	storageDrivers = make(map[string]NewStorageFunc)

	storageLocker sync.Mutex
)

func RegisterStorage(dirverName string, newFunc NewStorageFunc) (err error) {
	if dirverName == "" {
		err = errors.New("redconf: driver name could not be empty")
		return
	}

	if newFunc == nil {
		err = errors.New("redconf: new storage func could not be nil")
		return
	}

	storageLocker.Lock()
	defer storageLocker.Unlock()

	if _, exist := storageDrivers[dirverName]; exist {
		err = errors.New("redconf: storage driver of " + dirverName + " already exist")
		return
	}

	storageDrivers[dirverName] = newFunc

	return
}

func CreateStorage(driverName string, opts Options) (storage Storage, err error) {
	if driverName == "" {
		err = errors.New("redconf: driver name could not be empty")
		return
	}

	var exist bool
	var newFunc NewStorageFunc

	if newFunc, exist = storageDrivers[driverName]; !exist {
		err = errors.New("redconf: storage driver of " + driverName + " not exist")
		return
	}

	if storage, err = newFunc(opts); err != nil {
		return
	}

	return
}
