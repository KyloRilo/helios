package test

import (
	"reflect"
	"testing"
)

func assert(t *testing.T, val interface{}, expected interface{}, msg string) {
	if val != expected {
		t.Errorf("%s => Val: %v, Expected: %v", msg, val, expected)
	}
}

func assertPanic(t *testing.T, fn interface{}, params ...interface{}) {
	fnVal := reflect.ValueOf(fn)
	if fnVal.Kind() != reflect.Func {
		t.Errorf("Function param is not of type func")
	}

	if fnVal.Type().NumIn() != len(params) {
		t.Errorf("Mismatch in arguments for function param")
	}

	in := make([]reflect.Value, len(params))
	for i, param := range params {
		in[i] = reflect.ValueOf(param)
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected Panic")
		}
	}()
	fnVal.Call(in)
}
