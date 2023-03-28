package box

import (
	"context"
	"flag"
	"fmt"
	"reflect"

	"github.com/daemtri/di"
	"github.com/daemtri/di/container"
)

var (
	errType    = reflect.TypeOf(func() error { return nil }).Out(0)
	stdCtxType = reflect.TypeOf(func(context.Context) {}).In(0)
	flagAdder  = reflect.TypeOf(func(interface{ AddFlags(fs *flag.FlagSet) }) {}).In(0)
)

func reflectType[T any]() reflect.Type {
	return reflect.TypeOf(new(T)).Elem()
}
func emptyValue[T any]() (x T) { return }

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

func OptionFunc[T, K any](fn func(ctx context.Context, option K) (T, error)) *di.InjectBuilder[T, K] {
	return di.Inject(fn)
}

func Inject[T any](fn any, opt any) Builder[T] {
	fnType := reflect.TypeOf(fn)

	// 判断fn合法性
	if fnType.Kind() != reflect.Func {
		panic("ProvideInject only supports function types")
	}
	if fnType.NumOut() != 2 {
		panic("provideInject must return two parameters: (T,error) or (X,error), where X implements T")
	}
	pTyp := reflectType[T]()
	if pTyp.Kind() == reflect.Interface {
		if !fnType.Out(0).Implements(pTyp) {
			panic(fmt.Errorf("ProvideInject return type %s not implemented %s", fnType.Out(0), pTyp))
		}
	} else if pTyp != fnType.Out(0) {
		panic(fmt.Errorf("ProvideInject return value type %s != %s", fnType.Out(0), pTyp))
	}
	if fnType.Out(1) != errType {
		panic(fmt.Errorf("the second return value of the ProvideInject function must be %s", errType))
	}

	ib := &injectBuilder[T]{
		fnType:  fnType,
		fnValue: reflect.ValueOf(fn),
	}

	var flagTyp reflect.Type
	// 查找flagSetProvider
	for i := 0; i < fnType.NumIn(); i++ {
		if isFlagSetProvider(fnType.In(i)) {
			ib.optionIndex = i
			flagTyp = fnType.In(i)
			break
		}
	}

	if flagTyp != nil {
		if opt != nil {
			ib.Option = opt
		} else {
			var o reflect.Value
			if flagTyp.Kind() == reflect.Pointer {
				o = reflect.New(flagTyp.Elem())
			} else {
				o = reflect.Zero(flagTyp)
			}
			ib.Option = o.Interface()
		}
	}
	return ib
}

type injectBuilder[T any] struct {
	Option any `flag:""`

	optionIndex int
	fnValue     reflect.Value
	fnType      reflect.Type
}

func (ib *injectBuilder[T]) Build(ctx context.Context) (T, error) {
	defer func() {
		if e := recover(); e != nil {
			t := reflectType[T]()
			panic(fmt.Errorf("build(%s): %s", t, e))
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
		v := ctx.Value(container.ContextKey).(container.Interface).Invoke(ctx, ib.fnType.In(i))
		inValues = append(inValues, reflect.ValueOf(v))
	}

	ret := ib.fnValue.Call(inValues)
	if ret[1].Interface() == nil {
		return ret[0].Interface().(T), nil
	}
	return emptyValue[T](), ret[1].Interface().(error)
}
