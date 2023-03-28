package di

import (
	"encoding"
	"flag"
	"fmt"
	"reflect"
	"strings"
)

func isFlag(fieldTyp reflect.StructField) bool {
	_, ok := fieldTyp.Tag.Lookup("flag")
	return ok
}

func parseFlagTag(tag reflect.StructTag) (name, def, usage string) {
	return tag.Get("flag"), tag.Get("default"), tag.Get("usage")
}

func isNestedFlagStruct(typ reflect.Type) bool {
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return false
	}
	for i := 0; i < typ.NumField(); i++ {
		if !typ.Field(i).IsExported() {
			continue
		}
		if _, ok := typ.Field(i).Tag.Lookup("flag"); ok {
			return true
		}
	}
	return false
}

type prefix struct {
	name  string
	usage string
}

func (p prefix) concat(name, usage string) prefix {
	if name == "" {
		name = p.name
	} else if p.name != "" {
		name = p.name + "-" + name
	}
	if usage == "" {
		usage = p.usage
	} else if p.usage != "" {
		usage = p.usage + usage
	}
	return prefix{
		name:  name,
		usage: usage,
	}
}

func parseNested(fs *flag.FlagSet, pfx prefix, fType reflect.Type, fValue reflect.Value) {
	switch fType.Kind() {
	case reflect.Struct:
		parseStruct(fs, pfx, fType, fValue)
	case reflect.Pointer:
		if fType.Elem().Kind() == reflect.Struct {
			if fValue.IsNil() {
				fValue.Set(reflect.New(fType.Elem()))
			}
			parseStruct(fs, pfx, fType.Elem(), fValue.Elem())
		}
	}
}

func parseStruct(fs *flag.FlagSet, pfx prefix, fType reflect.Type, fValue reflect.Value) {
	fieldNum := fType.NumField()
	for i := 0; i < fieldNum; i++ {
		if !fType.Field(i).IsExported() || !isFlag(fType.Field(i)) {
			continue
		}
		name, def, usage := parseFlagTag(fType.Field(i).Tag)

		// 当前字段是一个flag，类型是一个Interface,并且这个Interface不是nil
		// 解析这个field的value的类型，如果是一个struct，那么继续解析
		if fType.Field(i).Type.Kind() == reflect.Interface {
			if fValue.Field(i).IsNil() {
				continue
			}
			fieldType := fValue.Field(i).Elem().Type()
			if !isNestedFlagStruct(fieldType) {
				continue
			}
			parseNested(fs, pfx.concat(name, usage), fieldType, fValue.Field(i).Elem())
			continue
		} else if isNestedFlagStruct(fType.Field(i).Type) {
			parseNested(fs, pfx.concat(name, usage), fType.Field(i).Type, fValue.Field(i))
			continue
		}

		if name == "" {
			// 如果没有指定flag的名称,则使用字段名
			name = strings.ToLower(fType.Field(i).Name)
		}
		tags := pfx.concat(name, usage)
		name, usage = tags.name, tags.usage

		if fType.Field(i).Type.Kind() == reflect.Pointer {
			panic(fmt.Errorf("flag parameter does not support pointer type,name=%s", name))
		}

		// 检查是否实现了
		// 		flag.Value接口
		//		encoding.Text{Marshaler,Unmarshaler}接口
		// 注: net.IP, time.Time均实现了该接口
		if fValue.Field(i).CanAddr() {
			v := fValue.Field(i).Addr().Interface()
			if vv, ok := v.(flag.Value); ok {
				if vv.String() == "" {
					if err := vv.Set(def); err != nil {
						panic(fmt.Errorf("failed to set default value for %s: %w", name, err))
					}
				}
				fs.Var(vv, name, usage)
			} else if vv, ok := v.(encoding.TextUnmarshaler); ok {
				if defValue, ok := v.(encoding.TextMarshaler); ok {
					current, err := defValue.MarshalText()
					if err != nil {
						panic(fmt.Errorf("failed to get current value of parameter: typ=%s,name=%s,err=%s", fValue.Field(i).Type(), name, err))
					}
					if len(current) == 0 && def != "" {
						if err := vv.UnmarshalText([]byte(def)); err != nil {
							panic(fmt.Errorf("parse default value error: typ=%s,name=%s,err=%s", fValue.Field(i).Type(), name, err))
						}
					}
					textVar(fs, vv, name, defValue, usage)
					continue
				}
			}
		}
		// 检查是否已经注册了reflect.Type类型处理器
		fn, ok := flagTypeBinds[fType.Field(i).Type]
		if !ok {
			// 检查是否已经注册了reflect.kind类别处理器
			fn, ok = flagKindBinds[fType.Field(i).Type.Kind()]
			if !ok {
				panic(fmt.Errorf("%s contains parameter type not supported: %s", fType, fType.Field(i).Type))
			}
		}
		if err := fn(fs, fValue.Field(i), name, def, usage); err != nil {
			panic(err)
		}
	}
}

type structFlagger struct {
	options any

	typ reflect.Type
	val reflect.Value
}

func newStructFlagger(options any) *structFlagger {
	return &structFlagger{
		options: options,
		typ:     reflect.TypeOf(options),
		val:     reflect.ValueOf(options),
	}
}

func (sf *structFlagger) AddFlags(fs *flag.FlagSet) {
	if add, ok := sf.options.(interface{ AddFlags(fs *flag.FlagSet) }); ok {
		add.AddFlags(fs)
		return
	}
	var fValue reflect.Value
	var fType reflect.Type

	if sf.typ.Kind() == reflect.Pointer {
		if !sf.val.IsValid() {
			sf.val = reflect.New(sf.typ.Elem())
		}
		fValue = sf.val.Elem()
		fType = sf.typ.Elem()
	} else {
		if !sf.val.IsValid() {
			sf.val = reflect.New(sf.typ).Elem()
		}
		fValue = sf.val
		fType = sf.typ
	}
	parseStruct(fs, prefix{}, fType, fValue)
}

func (sf *structFlagger) ValidateFlags() error {
	if optImpl, ok := sf.options.(interface{ ValidateFlags() error }); ok {
		return optImpl.ValidateFlags()
	}

	return validateFunc(sf.val.Interface())
}
