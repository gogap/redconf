package redconf

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type convFunc func(typ reflect.Type, value interface{}) (v interface{}, err error)

var (
	convFuncs map[reflect.Kind]convFunc = make(map[reflect.Kind]convFunc)

	defaultValue map[reflect.Kind]interface{} = make(map[reflect.Kind]interface{})
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

func convValue(typ reflect.Type, value interface{}) (v interface{}, err error) {
	if fn, exist := convFuncs[typ.Kind()]; exist {
		return fn(typ, value)
	}
	err = fmt.Errorf("could not conv Kind of %#v", typ.Kind())
	return
}

func convIntValue(typ reflect.Type, value interface{}) (v interface{}, err error) {
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

func convUintValue(typ reflect.Type, value interface{}) (v interface{}, err error) {
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

func convFloatValue(typ reflect.Type, value interface{}) (v interface{}, err error) {
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

func convBoolValue(typ reflect.Type, value interface{}) (v interface{}, err error) {
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

func convStringValue(typ reflect.Type, value interface{}) (v interface{}, err error) {
	if value == nil {
		v = getZeroValue(typ)
		return
	}

	if value == nil {
		v = getZeroValue(typ)
		return
	}

	strV := fmt.Sprintf("%s", value)
	strV = strings.TrimSpace(strV)

	v = strV

	return
}

func convSliceValue(typ reflect.Type, value interface{}) (v interface{}, err error) {
	if value == nil {
		v = getZeroValue(typ)
		return
	}

	strV := fmt.Sprintf("%s", value)
	strV = strings.TrimSpace(strV)

	strVs := strings.Split(strV, ",")

	if strVs == nil || len(strVs) == 0 {
		v = reflect.MakeSlice(typ, 0, 0).Interface()
		return
	}

	iVs := reflect.MakeSlice(typ, 0, len(strVs))
	tmpIvs := reflect.MakeSlice(typ, 1, 1)

	oneVType := tmpIvs.Index(0).Type()

	for i := 0; i < len(strVs); i++ {
		var oneV interface{}
		if oneV, err = convValue(oneVType, strVs[i]); err != nil {
			return
		}

		iVs = reflect.Append(iVs, reflect.ValueOf(oneV))
	}

	v = iVs.Interface()

	return
}
