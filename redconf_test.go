package redconf

import (
	"testing"
)

type TestBConfig struct {
	Field1     []int32
	Field2     string
	FieldArray string
}

type TestAConfig struct {
	Field1  string
	Field2  string
	Config3 TestBConfig
}

func TestRedConfStructParse(t *testing.T) {

	var monitor Monitor
	var storage Storage
	var opts Options
	var err error

	namespace := "NS"
	channel := "ONCHANGED"

	opts = Options{
		"address":  "localhost:6379",
		"password": "",
		"db":       0,
		"idle":     10,
		"channel":  channel,
	}

	if monitor, err = CreateMonitor("redis", opts); err != nil {
		t.Error(err)
		return
	}

	if storage, err = CreateStorage("redis", opts); err != nil {
		t.Error(err)
		return
	}

	var redConf *RedConf

	if redConf, err = New(namespace, storage, monitor); err != nil {
		return
	}

	conf := TestAConfig{}

	exceptedKeys := map[string]bool{
		"TestAConfig:Field1":             true,
		"TestAConfig:Field2":             true,
		"TestAConfig:Config3:Field1":     true,
		"TestAConfig:Config3:Field2":     true,
		"TestAConfig:Config3:FieldArray": true,
	}

	var fields []Field
	var watchConfig *WatchingConfig

	if watchConfig, err = NewWatchingConfig(&conf); err != nil {
		t.Error(err)
		return
	}

	redConf.WatchWithConfig(watchConfig)

	fields = watchConfig.Fields()

	if len(exceptedKeys) != len(fields) {
		t.Errorf("%#v\n", fields)
		return
	}

	for i := 0; i < len(exceptedKeys); i++ {
		if _, exist := exceptedKeys[fields[i].String()]; !exist {
			t.Errorf("err:%s\n%s\n", "fields not exist", fields[i])
			return
		}
	}
}
