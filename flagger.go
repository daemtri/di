package di

import (
	"encoding"
	"flag"
	"fmt"
	"reflect"
)

func isFlag(fieldTyp reflect.StructField) bool {
	_, ok := fieldTyp.Tag.Lookup("flag")
	return ok
}

func parseFlagTag(field reflect.StructField) (name, def, usage string) {
	return field.Tag.Get("flag"), field.Tag.Get("default"), field.Tag.Get("usage")
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
		name, defaultValue, usage := parseFlagTag(fType.Field(i))

		// The field is a flag, the type is an interface, and the interface is not nil.
		// Parse the value type of this field, if it is a struct, continue parsing.
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

		// the field name is empty string, and the field is not a nested flag, skip it.
		if name == "" {
			continue
		}
		full := pfx.concat(name, usage)
		if fType.Field(i).Type.Kind() == reflect.Pointer {
			panic(fmt.Errorf("flag parameter does not support pointer type,name=%s", full.name))
		}

		// Check that the value implements
		//		flag.Value
		//		encoding.Text{Marshaler,Unmarshaler}
		// Note: net.IP and time.Time implement both interfaces.
		if fValue.Field(i).CanAddr() {
			v := fValue.Field(i).Addr().Interface()
			if vv, ok := v.(flag.Value); ok {
				if vv.String() == "" {
					if err := vv.Set(defaultValue); err != nil {
						panic(fmt.Errorf("failed to set default value for %s: %w", full.name, err))
					}
				}
				fs.Var(vv, full.name, full.usage)
			} else if vv, ok := v.(encoding.TextUnmarshaler); ok {
				if defValue, ok := v.(encoding.TextMarshaler); ok {
					current, err := defValue.MarshalText()
					if err != nil {
						panic(fmt.Errorf("failed to get current value of parameter: typ=%s,name=%s,err=%s", fValue.Field(i).Type(), full.name, err))
					}
					if len(current) == 0 && defaultValue != "" {
						if err := vv.UnmarshalText([]byte(defaultValue)); err != nil {
							panic(fmt.Errorf("parse default value error: typ=%s,name=%s,err=%s", fValue.Field(i).Type(), full.name, err))
						}
					}
					fs.TextVar(vv, full.name, defValue, full.usage)
					continue
				}
			}
		}
		// check if we already have a handler for this reflect.Type
		fn, ok := flagTypeBinds[fType.Field(i).Type]
		if !ok {
			// Check if we have a registered reflect.Kind handler
			fn, ok = flagKindBinds[fType.Field(i).Type.Kind()]
			if !ok {
				panic(fmt.Errorf("%s contains parameter type not supported: %s", fType, fType.Field(i).Type))
			}
		}
		if err := fn(fs, fValue.Field(i), full.name, defaultValue, full.usage); err != nil {
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
	return nil
}
