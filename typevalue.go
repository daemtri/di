package di

import (
	"fmt"
	"reflect"
)

func reflectType[T any]() reflect.Type {
	return reflect.TypeOf(new(T)).Elem()
}

func emptyValue[T any]() (x T) { return }

func reflectTypeString(typ reflect.Type) string {
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
		return fmt.Sprintf("*%s.%s", typ.PkgPath(), typ.Name())
	}
	return fmt.Sprintf("%s.%s", typ.PkgPath(), typ.Name())
}

func reflectNew[T any]() T {
	typ := reflectType[T]()
	switch typ.Kind() {
	case reflect.Interface:
		return emptyValue[T]()
	case reflect.Pointer:
		return reflect.New(typ.Elem()).Interface().(T)
	default:
		return reflect.New(typ).Elem().Interface().(T)
	}
}
