package validate

import (
	"reflect"
)

func isFlag(fieldTyp reflect.StructField) bool {
	_, ok := fieldTyp.Tag.Lookup("flag")
	return ok
}

func parserTag(field reflect.StructField) (name, validate string) {
	name, validate = field.Tag.Get("flag"), field.Tag.Get("validate")
	return
}

func parseStruct(store map[string]string, prefix string, fType reflect.Type, fValue reflect.Value) {
	if fType.Kind() == reflect.Ptr {
		fType = fType.Elem()
		fValue = fValue.Elem()
	}
	for i := 0; i < fType.NumField(); i++ {
		field := fType.Field(i)
		if !field.IsExported() || !isFlag(field) {
			continue
		}

		fieldType := field.Type
		fieldValue := fValue.Field(i)
		if fieldType.Kind() == reflect.Interface {
			if fieldValue.IsNil() {
				continue
			}
			fieldValue = fieldValue.Elem()
			fieldType = fieldValue.Type()
		}
		name, validate := parserTag(field)
		if prefix != "" {
			if name == "" {
				name = prefix
			} else {
				name = prefix + "-" + name
			}
		}
		if fieldType.Kind() == reflect.Struct || (fieldType.Kind() == reflect.Ptr && fieldType.Elem().Kind() == reflect.Struct) {
			parseStruct(store, name, fieldType, fieldValue)
			continue
		}
		if validate != "" {
			store[name] = validate
		}
	}
}

func ParseValidateString(prefix string, v any) map[string]string {
	store := make(map[string]string)
	parseStruct(store, prefix, reflect.TypeOf(v), reflect.ValueOf(v))
	return store
}
