package box

import (
	"fmt"
	"reflect"

	"github.com/daemtri/di"
)

type wrapBuilder[T any] struct {
	Builder[T]
	selects map[reflect.Type]string
}

func (wb *wrapBuilder[T]) Build(ctx Context) (T, error) {
	return wb.Builder.Build(&reflectSelectedContext{
		Context: ctx,
		selects: wb.selects,
	})
}

func (wb *wrapBuilder[T]) Retrofit() error {
	if r, ok := wb.Builder.(Retrofiter); ok {
		return r.Retrofit()
	}
	return nil
}

type reflectSelectedContext struct {
	Context
	selects map[reflect.Type]string
}

func (rsc *reflectSelectedContext) SelectContext(p reflect.Type) Context {
	if name, ok := rsc.selects[p]; ok {
		return rsc.Context.Select(name)
	}
	return rsc.Context
}

func provide[T any](b Builder[T], opts ...Options) {
	if nfsIsParsed {
		panic(fmt.Errorf("不能在Build之后再执行Provide: %T", b))
	}
	opt := newOptions()
	for i := range opts {
		opts[i].apply(opt)
	}
	reg := defaultRegistrar
	if opt.name != "" {
		reg = defaultRegistrar.Named(opt.name)
	}
	if opt.override {
		reg = reg.Override()
	}
	bl := &wrapBuilder[T]{
		Builder: b,
		selects: opt.selects,
	}
	di.Provide[T](reg, bl).AddFlags(nfs.FlagSet(opt.flagSetPrefix))
}

func ProvideInstance[T any](value T, opts ...Options) {
	provide(Instance(value), opts...)
}

func ProvideFunc[T any](fn func(ctx Context) (T, error), opts ...Options) {
	provide[T](di.Func(fn), opts...)
}

func ProvideInject[T any](fn any, opts ...Options) {
	Provide[T](Inject[T](fn, nil), opts...)
}

// Provide 实现智能提供数据和注入数据的功能
// fn函数必须返回 (T,error) 或者 (X, error),X 实现了T接口
func Provide[T any](fn any, opts ...Options) {
	if b, ok := fn.(Builder[T]); ok {
		provide(b, opts...)
		return
	}
	if bb, ok := fn.(BBuilder[T]); ok {
		provide[T](&bBuilderFunc[T]{
			BBuilder: bb,
		}, opts...)
		return
	}
	if f, ok := fn.(func(ctx Context) (T, error)); ok {
		ProvideFunc(f, opts...)
		return
	}
	if x, ok := fn.(T); ok {
		ProvideInstance(x, opts...)
		return
	}
	rtp := reflect.TypeOf(fn)
	if rtp.Kind() == reflect.Func {
		ProvideInject[T](fn, opts...)
		return
	}
}
