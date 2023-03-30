package validate

import (
	"reflect"
	"testing"
)

func Test_ParseValidateString(t *testing.T) {
	type A struct {
		Name string `flag:"name" validate:"required"`
	}
	type B struct {
		Name string `flag:"name" validate:"required"`
		A    `flag:"a"`
	}
	type C struct {
		Option any `flag:""`
	}
	ret := ParseValidateString("c", &C{Option: &B{}})
	if !reflect.DeepEqual(ret, map[string]string{
		"c-a-name": "required",
		"c-name":   "required",
	}) {
		t.Fatal("test failed", ret)
	}
}
