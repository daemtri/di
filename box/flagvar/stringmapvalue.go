package flagvar

import (
	"reflect"
	"strings"
)

func StringStringMapTo[T BaseType](in map[string]string) (map[string]T, error) {
	rtn := make(map[string]T, len(in))
	kind := reflect.TypeOf(rtn).Elem().Kind()
	for key := range in {
		n, err := StringToKind[kind](in[key])
		if err != nil {
			return nil, err
		}
		rtn[key] = n.(T)
	}
	return rtn, nil
}

func ToStringStringMap[T BaseType](in map[string]T) map[string]string {
	rtn := make(map[string]string, len(in))
	kind := reflect.TypeOf(in).Elem().Kind()
	for key := range in {
		rtn[key] = KindToString[kind](in[key])
	}
	return rtn
}

func StringStringMapToStringSlice(in map[string]string) []string {
	rtn := make([]string, 0, len(in))
	for key := range in {
		rtn = append(rtn, key+"="+in[key])
	}
	return rtn
}

func StringSliceToStringStringMap(in []string) map[string]string {
	rtn := make(map[string]string, len(in))
	for i := range in {
		ss := strings.SplitN(in[i], "=", 2)
		key := ss[0]
		val := ""
		if len(ss) > 1 {
			val = ss[1]
		}
		rtn[key] = val
	}
	return rtn
}

type StringMapValue[T BaseType] struct {
	value   *map[string]T
	changed bool
}

func StringMap[T BaseType](p *map[string]T) *StringMapValue[T] {
	smv := new(StringMapValue[T])
	smv.value = p
	return smv
}

func (s *StringMapValue[T]) String() string {
	// flag包判断isZeroValue时会通过反射创建stringSliceValue然后调用String方法
	if s == nil || s.value == nil {
		return ""
	}
	str, _ := writeAsCSV(StringStringMapToStringSlice(ToStringStringMap(*s.value)))
	return "[" + str + "]"
}

func (s *StringMapValue[T]) Get() any {
	return *s.value
}

func (s *StringMapValue[T]) Set(val string) error {
	v, err := readAsCSV(val)
	if err != nil {
		return err
	}
	iv, err := StringStringMapTo[T](StringSliceToStringStringMap(v))
	if err != nil {
		return err
	}
	if !s.changed {
		*s.value = iv
	} else {
		for k := range iv {
			(*s.value)[k] = iv[k]
		}
	}
	s.changed = true
	return nil
}
