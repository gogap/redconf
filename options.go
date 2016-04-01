package redconf

import (
	"bytes"
	"encoding/json"
	"reflect"
)

type Options map[string]interface{}

func (p Options) Get(name string, v interface{}) (exist bool) {

	var opt interface{}
	if opt, exist = p[name]; exist {
		valOpt := reflect.ValueOf(opt)
		valVal := reflect.ValueOf(v)

		if valVal.Kind() == reflect.Ptr {
			valVal = valVal.Elem()
		}

		valVal.Set(valOpt)
	}

	return
}

func (p Options) ToObject(v interface{}) (err error) {
	var data []byte
	if data, err = json.Marshal(p); err != nil {
		return
	}

	buf := bytes.NewBuffer(data)
	decoder := json.NewDecoder(buf)
	decoder.UseNumber()

	if err = decoder.Decode(v); err != nil {
		return
	}

	return
}
