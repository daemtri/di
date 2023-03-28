package wire

import (
	"context"
	"flag"
	"fmt"
	"reflect"

	"github.com/daemtri/di/container"
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

func newAnyFunctionBuilder(fn any) *anyFunctionBuilder {
	fnType := reflect.TypeOf(fn)
	// 判断fn合法性
	if fnType.Kind() != reflect.Func {
		panic("ProvideInject only supports function types")
	}
	if fnType.NumOut() != 1 {
		panic("provideInject must return one parameters: T")
	}
	targetType := fnType.Out(0)
	ib := &anyFunctionBuilder{
		targetType: targetType,
		fnType:     fnType,
		fnValue:    reflect.ValueOf(fn),
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

type anyFunctionBuilder struct {
	Option any `flag:""`

	targetType  reflect.Type
	optionIndex int
	fnValue     reflect.Value
	fnType      reflect.Type
}

func (b *anyFunctionBuilder) Build(ctx context.Context) (any, error) {
	defer func() {
		if e := recover(); e != nil {
			panic(fmt.Errorf("build(%s): %s", b.targetType, e))
		}
	}()
	inValues := make([]reflect.Value, 0, b.fnType.NumIn())
	for i := 0; i < b.fnType.NumIn(); i++ {
		if i == b.optionIndex && b.Option != nil {
			inValues = append(inValues, reflect.ValueOf(b.Option))
			continue
		}
		if b.fnType.In(i) == stdCtxType {
			inValues = append(inValues, reflect.ValueOf(ctx))
			continue
		}
		v := ctx.Value(container.ContextKey).(container.Interface).Invoke(ctx, b.fnType.In(i))
		inValues = append(inValues, reflect.ValueOf(v))
	}

	ret := b.fnValue.Call(inValues)
	return ret[0].Interface(), nil
}
