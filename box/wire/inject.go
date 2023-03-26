package wire

import (
	"context"
	"flag"
	"fmt"
	"reflect"
)

var (
	stdCtxType = reflect.TypeOf(func(context.Context) {}).In(0)
	flagAdder  = reflect.TypeOf(func(interface{ AddFlags(fs *flag.FlagSet) }) {}).In(0)
)

// isFlagSetProvider 判断一个类型是否是flag
//
//	type Options struct {
//		Addr string `flag:"addr,127.0.0.1,地址"`
//		User string `flag:"user,,用户名"`
//		Password string `flag:"password,,密码"`
//	}
func isFlagSetProvider(v reflect.Type) bool {
	if v.Implements(flagAdder) {
		return true
	}
	lv := v
	if lv.Kind() == reflect.Pointer {
		lv = lv.Elem()
	}
	if lv.Kind() != reflect.Struct {
		return false
	}
	if lv.NumField() < 0 {
		return false
	}
	for i := 0; i < lv.NumField(); i++ {
		if !lv.Field(i).IsExported() {
			continue
		}
		if _, ok := lv.Field(i).Tag.Lookup("flag"); ok {
			return true
		}
	}
	return false
}

func Inject(fn any) *injectBuilder {
	fnType := reflect.TypeOf(fn)

	// 判断fn合法性
	if fnType.Kind() != reflect.Func {
		panic("ProvideInject只支持函数类型")
	}
	if fnType.NumOut() < 1 {
		panic("ProvideInject 函数必须返回1个参数")
	}
	pTyp := fnType.Out(0)

	ib := &injectBuilder{
		pType:   pTyp,
		fnType:  fnType,
		fnValue: reflect.ValueOf(fn),
	}

	// 查找flagSetProvider
	for i := 0; i < fnType.NumIn(); i++ {
		if isFlagSetProvider(fnType.In(i)) {
			ib.optionIndex = i
			flagTyp := fnType.In(i)
			var o reflect.Value
			if flagTyp.Kind() == reflect.Pointer {
				o = reflect.New(flagTyp.Elem())
			} else {
				o = reflect.Zero(flagTyp)
			}
			ib.Option = o.Interface()
			break
		}
	}
	return ib
}

type injectBuilder struct {
	Option any `flag:",nested"`

	pType       reflect.Type
	optionIndex int
	fnValue     reflect.Value
	fnType      reflect.Type
}

type reflectBuilder interface {
	Exists(p reflect.Type) bool
	Must(p reflect.Type) any
}

func (ib *injectBuilder) Build(ctx context.Context) (any, error) {
	defer func() {
		if e := recover(); e != nil {
			panic(fmt.Errorf("build(%s): %s", ib.pType, e))
		}
	}()
	inValues := make([]reflect.Value, 0, ib.fnType.NumIn())
	for i := 0; i < ib.fnType.NumIn(); i++ {
		if i == ib.optionIndex && ib.Option != nil {
			inValues = append(inValues, reflect.ValueOf(ib.Option))
			continue
		}
		if ib.fnType.In(i) == stdCtxType {
			inValues = append(inValues, reflect.ValueOf(ctx))
			continue
		}
		v := ctx.(reflectBuilder).Must(ib.fnType.In(i))
		inValues = append(inValues, reflect.ValueOf(v))
	}

	ret := ib.fnValue.Call(inValues)
	return ret[0].Interface(), nil
}
