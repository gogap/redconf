package redconf

import (
	"reflect"
	"sync"
)

type Field struct {
	name        string
	parents     []string
	structField reflect.StructField
	topVal      interface{}
	level       int
	str         string
	valLock     sync.Mutex
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
	p.valLock.Lock()
	defer p.valLock.Unlock()

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

	return
}

func (p *Field) Value() (currentVal interface{}) {
	val := reflect.ValueOf(p.topVal).Elem()

	for _, parent := range p.parents {
		val = val.FieldByName(parent)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
	}

	currentVal = val.FieldByName(p.name).Interface()

	return
}
