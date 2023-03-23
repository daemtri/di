//go:build !go1.19

package di

import (
	"encoding"
	"flag"
	"fmt"
	"reflect"
)

func textVar(fs *flag.FlagSet, p encoding.TextUnmarshaler, name string, value encoding.TextMarshaler, usage string) {
	fs.Var(newTextValue(value, p), name, usage)
}

// -- encoding.TextUnmarshaler Value
type textValue struct{ p encoding.TextUnmarshaler }

func newTextValue(val encoding.TextMarshaler, p encoding.TextUnmarshaler) textValue {
	ptrVal := reflect.ValueOf(p)
	if ptrVal.Kind() != reflect.Ptr {
		panic("variable value type must be a pointer")
	}
	defVal := reflect.ValueOf(val)
	if defVal.Kind() == reflect.Ptr {
		defVal = defVal.Elem()
	}
	if defVal.Type() != ptrVal.Type().Elem() {
		panic(fmt.Sprintf("default type does not match variable type: %v != %v", defVal.Type(), ptrVal.Type().Elem()))
	}
	ptrVal.Elem().Set(defVal)
	return textValue{p}
}

func (v textValue) Set(s string) error {
	return v.p.UnmarshalText([]byte(s))
}

func (v textValue) Get() interface{} {
	return v.p
}

func (v textValue) String() string {
	if m, ok := v.p.(encoding.TextMarshaler); ok {
		if b, err := m.MarshalText(); err == nil {
			return string(b)
		}
	}
	return ""
}
