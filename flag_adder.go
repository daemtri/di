package di

import (
	"encoding"
	"flag"
	"fmt"
	"reflect"
	"strings"
)

func parseFlagTag(tag reflect.StructTag) (name, def, usage string, isNested bool) {
	names := tag.Get("flag")
	namesArr := strings.SplitN(names, ",", 2)
	if len(namesArr) == 2 {
		isNested = namesArr[1] == "nested"
	}
	return namesArr[0], tag.Get("default"), tag.Get("usage"), isNested
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
		if !fType.Field(i).IsExported() {
			continue
		}
		name, def, usage, isNested := parseFlagTag(fType.Field(i).Tag)

		if fType.Field(i).Anonymous || isNested {
			npfx := pfx.concat(name, usage)
			if fType.Field(i).Type.Kind() == reflect.Interface {
				if !fValue.Field(i).IsNil() {
					parseNested(fs, npfx, fValue.Field(i).Elem().Type(), fValue.Field(i).Elem())
				}
			} else {
				parseNested(fs, npfx, fType.Field(i).Type, fValue.Field(i))
			}
			continue
		}

		if name == "" {
			continue
		}
		tags := pfx.concat(name, usage)
		name, usage = tags.name, tags.usage

		if fType.Field(i).Type.Kind() == reflect.Pointer {
			panic(fmt.Errorf("flag参数不支持指针类型,name=%s", name))
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
						panic(fmt.Errorf("设置%s默认值失败: %w", name, err))
					}
				}
				fs.Var(vv, name, usage)
			} else if vv, ok := v.(encoding.TextUnmarshaler); ok {
				if defValue, ok := v.(encoding.TextMarshaler); ok {
					current, err := defValue.MarshalText()
					if err != nil {
						panic(fmt.Errorf("获取参数当前值失败: typ=%s,name=%s,err=%s", fValue.Field(i).Type(), name, err))
					}
					if len(current) == 0 && def != "" {
						if err := vv.UnmarshalText([]byte(def)); err != nil {
							panic(fmt.Errorf("解析参数默认值错误: typ=%s,name=%s,err=%s", fValue.Field(i).Type(), name, err))
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
				panic(fmt.Errorf("%s包含暂不支持的参数类型: %s", fType, fType.Field(i).Type))
			}
		}
		if err := fn(fs, fValue.Field(i), name, def, usage); err != nil {
			panic(err)
		}
	}
}

type structFlagger struct {
	builder any

	typ reflect.Type
	val reflect.Value
}

func (sf *structFlagger) AddFlags(fs *flag.FlagSet) {
	if add, ok := sf.builder.(interface{ AddFlags(fs *flag.FlagSet) }); ok {
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
	if optImpl, ok := sf.builder.(interface{ ValidateFlags() error }); ok {
		return optImpl.ValidateFlags()
	}

	return validateFunc(sf.val.Interface())
}
