package redconf

import (
	"errors"
	"reflect"
	"strings"
	"sync"
)

type Field struct {
	name        string
	parents     []string
	structField reflect.StructField
	topVal      interface{}
	level       int
	str         string
}

func (p *Field) Name() string {
	return p.name
}

func (p *Field) Type() reflect.Type {
	return p.structField.Type
}

func (p *Field) Kind() reflect.Kind {
	return p.structField.Type.Kind()
}

func (p *Field) Level() int {
	return p.level
}

func (p *Field) String() string {
	return p.str
}

func (p *Field) Parents() []string {
	return p.parents
}

func (p *Field) set(v interface{}) {

	val := reflect.ValueOf(p.topVal).Elem()

	for _, parent := range p.parents {
		val = val.FieldByName(parent)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
	}

	val = val.FieldByName(p.name)

	if val.Kind() == reflect.Ptr && reflect.Zero(val.Type()) != v {
		val.Set(reflect.ValueOf(v))
	} else {
		val.Set(reflect.ValueOf(v))
	}

}

type WatchingConfig struct {
	name  string
	value interface{}

	fields   []*Field
	initOnce sync.Once
}

func NewWatchingConfig(v interface{}, name ...string) (wConf *WatchingConfig, err error) {
	if v == nil {
		err = errors.New("redconf: watching config value could not be nil")
		return
	}

	val := reflect.ValueOf(v)

	if !val.IsValid() {
		err = errors.New("redconf: reflect value error")
		return
	}

	if val.Kind() != reflect.Ptr {
		err = errors.New("redconf: watching config value should be a pointer to config struct")
		return
	}

	val = val.Elem()

	if val.Kind() != reflect.Struct {
		err = errors.New("redconf: watching config value should be kind of struct")
		return
	}

	t := reflect.Indirect(val).Type()

	confName := ""
	if name == nil {
		confName = t.Name()
	} else if len(name) > 0 && name[0] != "" {
		confName = name[0]
	}

	conf := &WatchingConfig{
		name:  confName,
		value: v,
	}

	conf.initOnce.Do(func() {
		if conf.fields, err = conf.initFields(v); err != nil {
			return
		}
	})

	wConf = conf

	return
}

func (p *WatchingConfig) Name() string {
	return p.name
}

func (p *WatchingConfig) Value(v interface{}) {
	return
}

func (p *WatchingConfig) Fields() []Field {
	var fs []Field

	for _, field := range p.fields {
		fs = append(fs, *field)
	}
	return fs
}

func (p *WatchingConfig) initFields(v interface{}) (fields []*Field, err error) {
	val := reflect.ValueOf(v)

	if !val.IsValid() {
		err = errors.New("redconf: reflect value error")
		return
	}

	if fields, err = p.getStructFields([]string{}, val, 0); err != nil {
		return
	}

	for _, field := range fields {
		field.topVal = v
	}

	return
}

func (p *WatchingConfig) getStructFields(parents []string, val reflect.Value, level int) (fields []*Field, err error) {

	var t reflect.Type
	var isSupport bool

	if val, t, isSupport = getRelValueAndType(val); !isSupport {
		return
	}

	if !val.IsValid() {
		err = errors.New("redconf: reflect value error")
		return
	}

	var tmpFields []*Field

	for i := 0; i < t.NumField(); i++ {
		vKind := val.Field(i).Kind()

		if vKind == reflect.Ptr {
			vKind = val.Field(i).Type().Elem().Kind()
		}

		switch vKind {
		case reflect.Struct:
			{
				var tFields []*Field
				var nextVal reflect.Value

				nextIsSupport := false

				if nextVal, _, nextIsSupport = getRelValueAndType(val.Field(i)); !nextIsSupport {
					return
				}

				if val.Field(i).Kind() == reflect.Ptr && val.Field(i).IsNil() {
					val.Field(i).Set(nextVal)
				}

				if tFields, err = p.getStructFields(append(parents, t.Field(i).Name), nextVal, level+1); err != nil {
					return
				}
				tmpFields = append(tmpFields, tFields...)
			}
		case reflect.Array, reflect.Slice,
			reflect.Bool, reflect.Map,
			reflect.Float32, reflect.Float64,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.String:

			tmpStrs := []string{p.name}
			tmpStrs = append(tmpStrs, parents...)
			tmpStrs = append(tmpStrs, t.Field(i).Name)

			field := &Field{
				name:        t.Field(i).Name,
				parents:     parents,
				level:       level,
				structField: t.Field(i),
				str:         strings.Join(tmpStrs, ":"),
			}

			tmpFields = append(tmpFields, field)
		}
	}

	fields = tmpFields
	return
}

func getRelValueAndType(val reflect.Value) (retV reflect.Value, retType reflect.Type, isSupport bool) {

	isSupport = true

	if val.Kind() == reflect.Ptr {
		if val.Type().Elem().Kind() == reflect.Struct {
			if val.IsNil() {
				retV = reflect.New(val.Type().Elem())
				retType = val.Type().Elem()
			} else {
				retV = val.Elem()
				retType = reflect.Indirect(val).Type()
			}
		} else {
			isSupport = false
		}
	} else if val.Kind() == reflect.Struct {
		retType = val.Type()
		retV = reflect.New(val.Type()).Elem()
	} else {
		isSupport = false
	}

	return
}
