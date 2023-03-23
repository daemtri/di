package flagvar

import (
	"reflect"
)

func StringSliceTo[T BaseType](in []string) ([]T, error) {
	rtn := make([]T, 0, len(in))
	kind := reflect.TypeOf(rtn).Elem().Kind()
	for i := range in {
		n, err := StringToKind[kind](in[i])
		if err != nil {
			return nil, err
		}
		rtn = append(rtn, n.(T))
	}
	return rtn, nil
}

func ToStringSlice[T BaseType](in []T) []string {
	rtn := make([]string, 0, len(in))
	kind := reflect.TypeOf(in).Elem().Kind()
	for i := range in {
		rtn = append(rtn, KindToString[kind](in[i]))
	}
	return rtn
}

// -- SliceValue Value
type SliceValue[T BaseType] struct {
	value   *[]T
	changed bool
}

func Slice[T BaseType](p *[]T, val ...T) *SliceValue[T] {
	ssv := new(SliceValue[T])
	ssv.value = p
	*ssv.value = val
	return ssv
}

func (s *SliceValue[T]) Set(val string) error {
	v, err := readAsCSV(val)
	if err != nil {
		return err
	}
	iv, err := StringSliceTo[T](v)
	if err != nil {
		return err
	}
	if !s.changed {
		*s.value = iv
	} else {
		*s.value = append(*s.value, iv...)
	}
	s.changed = true
	return nil
}

func (s *SliceValue[T]) String() string {
	// flag包判断isZeroValue时会通过反射创建stringSliceValue然后调用String方法
	if s == nil || s.value == nil {
		return ""
	}
	str, _ := writeAsCSV(ToStringSlice(*s.value))
	return "[" + str + "]"
}

func (s *SliceValue[T]) Get() any {
	return *s.value
}
