package box

import (
	"context"
	"fmt"
	"reflect"

	"github.com/daemtri/di"
)

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
		reg = reg.Named(opt.name)
	}
	if opt.override {
		reg = reg.Override()
	}

	c := di.Provide[T](reg, b).AddFlags(nfs.FlagSet(opt.flagSetPrefix))
	if opt.selects != nil {
		c.Designate(opt.selects...)
	}
}

func ProvideFunc[T any](fn func(ctx context.Context) (T, error), opts ...Options) {
	provide[T](di.Func(fn), opts...)
}

func provideInject[T any](fn any, opts ...Options) {
	provide(Inject[T](fn, nil), opts...)
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
	if f, ok := fn.(func(ctx context.Context) (T, error)); ok {
		ProvideFunc(f, opts...)
		return
	}
	rtp := reflect.TypeOf(fn)
	if rtp.Kind() == reflect.Func {
		provideInject[T](fn, opts...)
		return
	}
}
