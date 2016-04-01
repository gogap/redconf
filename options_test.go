package redconf

import (
	"testing"
)

type optTestStruct struct {
	A int
	B string
}

func TestRedConfOptions(t *testing.T) {

	opts := Options{"A": 123, "B": "456"}

	var a int
	var b string

	opts.Get("A", &a)
	opts.Get("B", &b)

	if a != 123 {
		t.Error("get value a failed")
		return
	}

	if b != "456" {
		t.Error("get value b failed")
		return
	}

	optSt := optTestStruct{}

	opts.ToObject(&optSt)

	if optSt.A != 123 || optSt.B != "456" {
		t.Error("to object failed")
		return
	}

	return
}
