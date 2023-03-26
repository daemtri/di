package di

import (
	"context"
	"fmt"
	"reflect"
)

func getTypeNameFromContext(ctx context.Context, typ reflect.Type) string {
	secs := getContext(ctx).requirer().constructor.selections
	if secs == nil {
		return ""
	}
	return secs[typ]
}

// Builder 定义了对象构建器
type Builder[T any] interface {
	Build(ctx context.Context) (T, error)
}

// Exists 断言某个类型已经Provided
func Exists[T any](ctx context.Context) bool {
	typ := reflectType[T]()
	s, ok := getContext(ctx).container().constructors[typ]
	if !ok {
		return false
	}
	return s.exists(getTypeNameFromContext(ctx, typ))
}

// Must 只能在BuildFactory中使用
func Must[T any](ctx context.Context) T {
	typ := reflectType[T]()
	v, err := getContext(ctx).container().build(ctx, typ, getTypeNameFromContext(ctx, typ))
	if err != nil {
		panic(fmt.Errorf("must 构建失败： %s", err))
	}
	return v.(T)
}

// MustAll 构建某个类型的所有注册对象
func MustAll[T any](ctx context.Context) map[string]T {
	typ := reflectType[T]()
	localCtx := getContext(ctx)
	s, ok := localCtx.container().constructors[typ]
	if !ok {
		panic(fmt.Errorf("类型: %s不存在", typ))
	}
	values := make(map[string]T, len(s.groups))
	for name := range s.groups {
		v, err := localCtx.container().build(ctx, typ, name)
		if err != nil {
			panic(err)
		}
		values[name] = v.(T)
	}
	return values
}

// Build 构建一个指定对象
func Build[T any](reg Registry, ctx context.Context) (T, error) {
	if err := reg.ValidateFlags(); err != nil {
		return emptyValue[T](), err
	}
	ctx2 := withContext(ctx, newBaseContext(reg.container))
	typ := reflectType[T]()
	v, err := reg.build(ctx2, typ, reg.name)
	if err != nil {
		return emptyValue[T](), err
	}
	return v.(T), err
}
