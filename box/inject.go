package box

import (
	"context"
	"flag"
	"fmt"
	"reflect"

	"github.com/daemtri/di"
)

var (
	errType    = reflect.TypeOf(func() error { return nil }).Out(0)
	ctxType    = reflect.TypeOf(func(Context) {}).In(0)
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

func OptionFunc[T, K any](fn func(ctx Context, option K) (T, error)) *di.InjectBuilder[T, K] {
	return di.Inject[T, K](fn)
}

func Inject[T any](fn any, opt any) Builder[T] {
	fnType := reflect.TypeOf(fn)

	// 判断fn合法性
	if fnType.Kind() != reflect.Func {
		panic("ProvideInject只支持函数类型")
	}
	if fnType.NumOut() != 2 {
		panic("ProvideInject 函数必须返回2个参数: (T,error) 或者 (X, error),X 实现了T接口")
	}
	pTyp := reflectType[T]()
	if pTyp.Kind() == reflect.Interface {
		if !fnType.Out(0).Implements(pTyp) {
			panic(fmt.Errorf("ProvideInject 函数返回值类型 %s 未实现 %s", fnType.Out(0), pTyp))
		}
	} else if pTyp != fnType.Out(0) {
		panic(fmt.Errorf("ProvideInject 函数返回值类型 %s != %s", fnType.Out(0), pTyp))
	}
	if fnType.Out(1) != errType {
		panic(fmt.Errorf("ProvideInject 函数第二个返回值必须为 %s", errType))
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
	Option any `flag:",nested"`

	optionIndex int
	fnValue     reflect.Value
	fnType      reflect.Type
}

type reflectBuilder interface {
	Exists(p reflect.Type) bool
	Must(p reflect.Type) any
}

func (ib *injectBuilder[T]) Build(ctx Context) (T, error) {
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
		if ib.fnType.In(i) == ctxType {
			inValues = append(inValues, reflect.ValueOf(ctx))
			continue
		}
		if ib.fnType.In(i) == stdCtxType {
			inValues = append(inValues, reflect.ValueOf(ctx.Unwrap()))
			continue
		}

		lCtx := ctx
		if st, ok := ctx.(*reflectSelectedContext); ok {
			lCtx = st.SelectContext(ib.fnType.In(i))
		}
		v := lCtx.(reflectBuilder).Must(ib.fnType.In(i))
		inValues = append(inValues, reflect.ValueOf(v))
	}

	ret := ib.fnValue.Call(inValues)
	if ret[1].Interface() == nil {
		return ret[0].Interface().(T), nil
	}
	return emptyValue[T](), ret[1].Interface().(error)
}

// instanceBuilder
type instanceBuilder[T any] struct {
	Instance T `flag:",nested"`
}

func Instance[T any](v T) Builder[T] {
	return &instanceBuilder[T]{
		Instance: v,
	}
}

func (b *instanceBuilder[T]) Build(ctx Context) (T, error) {
	refTyp := reflect.TypeOf(b.Instance)
	refVal := reflect.ValueOf(b.Instance)

	if refTyp.Kind() == reflect.Pointer {
		refTyp = refTyp.Elem()
		refVal = refVal.Elem()
	}

	if refTyp.Kind() != reflect.Struct {
		return b.Instance, nil
	}

	for i := 0; i < refTyp.NumField(); i++ {
		if !refVal.Field(i).CanSet() {
			continue
		}
		injectType, ok := refTyp.Field(i).Tag.Lookup("inject")
		if !ok {
			continue
		}
		if injectType == "must" {
			lCtx := ctx
			if st, ok := ctx.(*reflectSelectedContext); ok {
				lCtx = st.SelectContext(refTyp.Field(i).Type)
			}
			v := lCtx.(reflectBuilder).Must(refTyp.Field(i).Type)
			refVal.Field(i).Set(reflect.ValueOf(v))
			continue
		}
		if injectType == "exists" {
			lCtx := ctx
			if st, ok := ctx.(*reflectSelectedContext); ok {
				lCtx = st.SelectContext(refTyp.Field(i).Type)
			}
			if lCtx.(reflectBuilder).Exists(refTyp.Field(i).Type) {
				v := lCtx.(reflectBuilder).Must(refTyp.Field(i).Type)
				refVal.Field(i).Set(reflect.ValueOf(v))
			}
			continue
		}
	}

	return b.Instance, nil
}
