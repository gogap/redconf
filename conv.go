package redconf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type convFunc func(typ reflect.Type, value interface{}, toPtr bool) (v interface{}, err error)

var (
	convFuncs = make(map[reflect.Kind]convFunc)

	defaultValue = make(map[reflect.Kind]interface{})
)

func init() {
	convFuncs[reflect.Bool] = convBoolValue
	convFuncs[reflect.Int] = convIntValue
	convFuncs[reflect.Int8] = convIntValue
	convFuncs[reflect.Int16] = convIntValue
	convFuncs[reflect.Int32] = convIntValue
	convFuncs[reflect.Int64] = convIntValue
	convFuncs[reflect.Uint] = convUintValue
	convFuncs[reflect.Uint8] = convUintValue
	convFuncs[reflect.Uint16] = convUintValue
	convFuncs[reflect.Uint32] = convUintValue
	convFuncs[reflect.Uint64] = convUintValue
	convFuncs[reflect.Float32] = convFloatValue
	convFuncs[reflect.Float64] = convFloatValue
	convFuncs[reflect.String] = convStringValue
	convFuncs[reflect.Slice] = convSliceValue
	convFuncs[reflect.Struct] = convStructValue
	convFuncs[reflect.Map] = convMapValue
	convFuncs[reflect.Ptr] = convPtrValue

	defaultValue[reflect.Bool] = false
	defaultValue[reflect.Int] = int(0)
	defaultValue[reflect.Int8] = int8(0)
	defaultValue[reflect.Int16] = int16(0)
	defaultValue[reflect.Int32] = int32(0)
	defaultValue[reflect.Int64] = int64(0)
	defaultValue[reflect.Uint] = uint(0)
	defaultValue[reflect.Uint8] = uint8(0)
	defaultValue[reflect.Uint16] = uint16(0)
	defaultValue[reflect.Uint32] = uint32(0)
	defaultValue[reflect.Uint64] = uint64(0)
	defaultValue[reflect.Float32] = float32(0)
	defaultValue[reflect.Float64] = float64(0)
	defaultValue[reflect.String] = ""
	defaultValue[reflect.Slice] = nil
}

func getZeroValue(typ reflect.Type) (v interface{}) {
	v, _ = defaultValue[typ.Kind()]
	return
}

func conv(typ reflect.Type, value interface{}) (v interface{}, err error) {
	return convValue(typ, value, false)
}

func convValue(typ reflect.Type, value interface{}, toPtr bool) (v interface{}, err error) {
	if fn, exist := convFuncs[typ.Kind()]; exist {
		return fn(typ, value, toPtr)
	}
	err = fmt.Errorf("could not conv Kind of %#v", typ.Kind())
	return
}

func convIntValue(typ reflect.Type, value interface{}, toPtr bool) (v interface{}, err error) {
	if value == nil {
		v = getZeroValue(typ)
		return
	}

	switch typ.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		{
			if value == nil {
				v = getZeroValue(typ)
				return
			}

			strV := fmt.Sprintf("%s", value)

			if strV == "" {
				v = getZeroValue(typ)
				return
			}

			var intV int64
			if intV, err = strconv.ParseInt(strV, 10, 64); err != nil {
				return
			}

			switch typ.Kind() {
			case reflect.Int:
				{
					v = int(intV)
				}
			case reflect.Int8:
				{
					v = int8(intV)
				}
			case reflect.Int16:
				{
					v = int16(intV)
				}
			case reflect.Int32:
				{
					v = int32(intV)
				}
			case reflect.Int64:
				{
					v = intV
				}
			}
		}
	default:
		err = fmt.Errorf("redconf: could not conv value %#v to int", value)
		return
	}
	return
}

func convUintValue(typ reflect.Type, value interface{}, toPtr bool) (v interface{}, err error) {
	if value == nil {
		v = getZeroValue(typ)
		return
	}

	switch typ.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		{
			if value == nil {
				v = getZeroValue(typ)
				return
			}

			strV := fmt.Sprintf("%s", value)
			var intV uint64
			if intV, err = strconv.ParseUint(strV, 10, 64); err != nil {
				return
			}

			switch typ.Kind() {
			case reflect.Int:
				{
					v = uint(intV)
				}
			case reflect.Int8:
				{
					v = uint8(intV)
				}
			case reflect.Int16:
				{
					v = uint16(intV)
				}
			case reflect.Int32:
				{
					v = uint32(intV)
				}
			case reflect.Int64:
				{
					v = intV
				}
			}
		}
	default:
		err = fmt.Errorf("redconf: could not conv value %#v to uint", value)
		return
	}
	return
}

func convFloatValue(typ reflect.Type, value interface{}, toPtr bool) (v interface{}, err error) {
	if value == nil {
		v = getZeroValue(typ)
		return
	}

	switch typ.Kind() {
	case reflect.Float32, reflect.Float64:
		{
			strV := fmt.Sprintf("%s", value)
			var floatV float64
			if floatV, err = strconv.ParseFloat(strV, 64); err != nil {
				return
			}

			switch typ.Kind() {
			case reflect.Float32:
				{
					v = float32(floatV)
				}
			case reflect.Float64:
				{
					v = floatV
				}
			}
		}
	default:
		err = fmt.Errorf("redconf: could not conv value %#v to float", value)
		return
	}
	return
}

func convBoolValue(typ reflect.Type, value interface{}, toPtr bool) (v interface{}, err error) {
	if value == nil {
		v = getZeroValue(typ)
		return
	}

	switch typ.Kind() {
	case reflect.Bool:
		{
			if value == nil {
				v = getZeroValue(typ)
				return
			}

			strV := fmt.Sprintf("%s", value)
			strV = strings.TrimSpace(strV)

			bv := false
			bv, _ = strconv.ParseBool(strV)

			v = bv
		}
	default:
		err = fmt.Errorf("redconf: could not conv value %#v to bool", value)
		return
	}
	return
}

func convStringValue(typ reflect.Type, value interface{}, toPtr bool) (v interface{}, err error) {
	if value == nil {
		v = getZeroValue(typ)
		return
	}

	strV := fmt.Sprintf("%s", value)
	strV = strings.TrimSpace(strV)

	if toPtr {
		v = &strV
	} else {
		v = strV
	}

	return
}

func convSliceValue(typ reflect.Type, value interface{}, toPtr bool) (v interface{}, err error) {
	if value == nil {
		v = getZeroValue(typ)
		return
	}

	tmpIvs := reflect.MakeSlice(typ, 1, 1)
	oneVType := tmpIvs.Index(0).Type()

	strV := fmt.Sprintf("%s", value)
	strV = strings.TrimSpace(strV)

	switch oneVType.Kind() {
	case reflect.Bool,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Float32,
		reflect.Float64,
		reflect.String:
		{

			strVs := strings.Split(strV, ",")

			if strV == "" || strVs == nil || len(strVs) == 0 {
				v = reflect.MakeSlice(typ, 0, 0).Interface()
				return
			}

			iVs := reflect.MakeSlice(typ, 0, len(strVs))

			for i := 0; i < len(strVs); i++ {
				var oneV interface{}
				if oneV, err = convValue(oneVType, strVs[i], false); err != nil {
					return
				}

				iVs = reflect.Append(iVs, reflect.ValueOf(oneV))
			}

			v = iVs.Interface()
		}
	case reflect.Map:
		{
			v, err = convMapValue(typ, value, false)
		}
	case reflect.Struct:
		{
			v, err = convStructSliceValue(typ, strV)
		}
	case reflect.Ptr:
		{
			if oneVType.Elem().Kind() == reflect.Struct {
				v, err = convStructSliceValue(typ, strV)
			}
		}
	}

	return
}

func convStructSliceValue(typ reflect.Type, str string) (v interface{}, err error) {
	if str == "" {
		return
	}

	sV := reflect.New(typ).Interface()

	buf := bytes.NewBufferString(str)

	decoder := json.NewDecoder(buf)
	decoder.UseNumber()

	if err = decoder.Decode(&sV); err != nil {
		return
	}

	v = reflect.Indirect(reflect.ValueOf(sV)).Interface()

	return
}

func convStructValue(typ reflect.Type, value interface{}, toPtr bool) (v interface{}, err error) {
	if value == nil {
		v = reflect.New(typ).Elem().Interface()
		return
	}

	strV := fmt.Sprintf("%s", value)
	strV = strings.TrimSpace(strV)

	if strV == "" {
		v = reflect.New(typ).Elem().Interface()
		return
	}

	buf := bytes.NewBufferString(strV)

	decoder := json.NewDecoder(buf)

	decoder.UseNumber()

	v = reflect.New(typ).Elem().Interface()

	if err = decoder.Decode(&v); err != nil {
		return
	}

	return
}

func convPtrValue(typ reflect.Type, value interface{}, toPtr bool) (v interface{}, err error) {

	if value == nil {
		return
	}

	var newV interface{}
	if newV, err = convValue(typ.Elem(), value, true); err != nil {
		return
	}

	v = newV

	return
}

func convMapValue(typ reflect.Type, value interface{}, toPtr bool) (v interface{}, err error) {
	if value == nil {
		return
	}

	strV := fmt.Sprintf("%s", value)
	strV = strings.TrimSpace(strV)

	if strV == "" {
		v = nil
		return
	}

	buf := bytes.NewBufferString(strV)

	decoder := json.NewDecoder(buf)
	decoder.UseNumber()

	newV := reflect.New(typ).Interface()

	if err = decoder.Decode(&newV); err != nil {
		return
	}

	v = reflect.Indirect(reflect.ValueOf(newV)).Interface()

	return
}
