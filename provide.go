package di

import (
	"fmt"
	"reflect"
)

func provideWithName[T any](reg Registry, b Builder[T], sf *structFlagger) Constructor {
	typ := reflectType[T]()
	var mb *multiConstructor
	if c, ok := reg.constructors[typ]; ok {
		mb, ok = c.(*multiConstructor)
		if !ok {
			panic(fmt.Errorf("provided %s 错误,已存在非命名对象", typ))
		}
		if mb.exists(reg.name) && !reg.override {
			panic(fmt.Errorf("provided %s:%s already exists", typ, reg.name))
		}
	} else {
		mb = &multiConstructor{cs: make(map[string]Constructor)}
	}

	c := &constructor{
		builder:           b,
		addFlags:          sf.AddFlags,
		validateFlagsFunc: sf.ValidateFlags,
		buildFunc: func(ctx Context) (any, error) {
			return b.Build(ctx)
		},
	}
	mb.cs[reg.name] = c
	mb.current = c
	reg.constructors[typ] = mb
	return mb
}

// Provide 向容器中注册构建器
func Provide[T any](reg Registry, b Builder[T]) Constructor {
	typ := reflectType[T]()
	sf := &structFlagger{
		builder: b,
		typ:     reflect.TypeOf(b),
		val:     reflect.ValueOf(b),
	}
	if reg.name != "" {
		return provideWithName(reg, b, sf)
	}
	if _, ok := reg.constructors[typ]; ok {
		if !reg.override {
			panic(fmt.Errorf("provided %s already exists", typ))
		}
	}

	c := &constructor{
		builder:           b,
		addFlags:          sf.AddFlags,
		validateFlagsFunc: sf.ValidateFlags,
		buildFunc: func(ctx Context) (any, error) {
			return b.Build(ctx)
		},
	}
	reg.constructors[typ] = c
	return c
}
