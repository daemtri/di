package flagvar

import (
	"fmt"
	"reflect"
	"strconv"
)

var (
	StringToKind = map[reflect.Kind]func(in string) (any, error){
		reflect.Int8: func(in string) (any, error) {
			n, err := strconv.ParseInt(in, 10, 8)
			return int8(n), err
		},
		reflect.Int16: func(in string) (any, error) {
			n, err := strconv.ParseInt(in, 10, 16)
			return int16(n), err
		},
		reflect.Int32: func(in string) (any, error) {
			n, err := strconv.ParseInt(in, 10, 32)
			return int32(n), err
		},
		reflect.Int64: func(in string) (any, error) {
			return strconv.ParseInt(in, 10, 64)
		},
		reflect.Int: func(in string) (any, error) {
			n, err := strconv.ParseInt(in, 10, strconv.IntSize)
			return int(n), err
		},
		reflect.Uint8: func(in string) (any, error) {
			n, err := strconv.ParseUint(in, 10, 8)
			return uint8(n), err
		},
		reflect.Uint16: func(in string) (any, error) {
			n, err := strconv.ParseUint(in, 10, 16)
			return uint16(n), err
		},
		reflect.Uint32: func(in string) (any, error) {
			n, err := strconv.ParseUint(in, 10, 32)
			return uint32(n), err
		},
		reflect.Uint64: func(in string) (any, error) {
			return strconv.ParseUint(in, 10, 64)
		},
		reflect.Uint: func(in string) (any, error) {
			n, err := strconv.ParseUint(in, 10, strconv.IntSize)
			return uint(n), err
		},
		reflect.Float32: func(in string) (any, error) {
			n, err := strconv.ParseFloat(in, 32)
			return float32(n), err
		},
		reflect.Float64: func(in string) (any, error) {
			return strconv.ParseFloat(in, 64)
		},
		reflect.String: func(in string) (any, error) {
			return in, nil
		},
		reflect.Bool: func(in string) (any, error) {
			return strconv.ParseBool(in)
		},
	}
	KindToString = map[reflect.Kind]func(x any) string{
		reflect.Int8: func(in any) string {
			return fmt.Sprintf("%d", in)
		},
		reflect.Int16: func(in any) string {
			return fmt.Sprintf("%d", in)
		},
		reflect.Int32: func(in any) string {
			return fmt.Sprintf("%d", in)
		},
		reflect.Int64: func(in any) string {
			return fmt.Sprintf("%d", in)
		},
		reflect.Int: func(in any) string {
			return fmt.Sprintf("%d", in)
		},
		reflect.Uint8: func(in any) string {
			return fmt.Sprintf("%d", in)
		},
		reflect.Uint16: func(in any) string {
			return fmt.Sprintf("%d", in)
		},
		reflect.Uint32: func(in any) string {
			return fmt.Sprintf("%d", in)
		},
		reflect.Uint64: func(in any) string {
			return fmt.Sprintf("%d", in)
		},
		reflect.Uint: func(in any) string {
			return fmt.Sprintf("%d", in)
		},
		reflect.Float32: func(in any) string {
			return fmt.Sprintf("%.2f", in)
		},
		reflect.Float64: func(in any) string {
			return fmt.Sprintf("%.2f", in)
		},
		reflect.String: func(in any) string {
			return in.(string)
		},
		reflect.Bool: func(in any) string {
			return strconv.FormatBool(in.(bool))
		},
	}
)

type BaseType interface {
	int8 | int16 | int32 | int64 | int | uint8 | uint16 | uint32 | uint64 | uint | float32 | float64 | string | bool
}

type BaseValue[T BaseType] struct {
	value *T
	kind  reflect.Kind
}

func Base[T BaseType](p *T, def T) *BaseValue[T] {
	bv := &BaseValue[T]{
		value: p,
		kind:  reflect.TypeOf(def).Kind(),
	}
	*bv.value = def
	return bv
}

func (s *BaseValue[T]) Set(val string) error {
	n, err := StringToKind[s.kind](val)
	if err != nil {
		return err
	}
	*s.value = n.(T)
	return nil
}

func (s *BaseValue[T]) String() string {
	// flag包判断isZeroValue时会通过反射创建,然后调用String方法
	if s == nil || s.value == nil {
		return ""
	}
	return KindToString[s.kind](*s.value)
}

func (s *BaseValue[T]) Get() any {
	return *s.value
}
