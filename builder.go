package di

import (
	"context"
	"fmt"
)

// Builder 定义了对象构建器
type Builder[T any] interface {
	Build(ctx Context) (T, error)
}

// Exists 断言某个类型已经Provided
// 如果ctx.Select了name，则判断name是否存在
func Exists[T any](ctx Context) bool {
	typ := reflectType[T]()
	s, ok := ctx.container().constructors[typ]
	if !ok {
		return false
	}
	if ctx.name() != "" {
		mtb, ok := s.(*multiConstructor)
		if ok {
			return mtb.exists(ctx.name())
		}
	}
	return true
}

// Must 只能在BuildFactory中使用
func Must[T any](ctx Context) T {
	typ := reflectType[T]()
	v, err := ctx.container().build(ctx, typ)
	if err != nil {
		panic(fmt.Errorf("must 构建失败： %s", err))
	}
	return v.(T)
}

// MustAll 构建某个类型的所有注册对象
func MustAll[T any](ctx Context) map[string]T {
	typ := reflectType[T]()
	s, ok := ctx.container().constructors[typ]
	if !ok {
		panic(fmt.Errorf("类型: %s不存在", typ))
	}
	mtb, ok := s.(*multiConstructor)
	if !ok {
		v, err := ctx.container().build(ctx, typ)
		if err != nil {
			panic(err)
		}
		return map[string]T{"": v.(T)}
	}
	values := make(map[string]T, len(mtb.cs))
	for name := range mtb.cs {
		v, err := ctx.container().build(ctx.Select(name), typ)
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
	var c Context = wrapContext(ctx, reg.container)
	if reg.name != "" {
		c = c.Select(reg.name)
	}
	typ := reflectType[T]()
	v, err := reg.build(c, typ)
	if err != nil {
		return emptyValue[T](), err
	}
	return v.(T), err
}
