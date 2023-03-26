package di

import (
	"fmt"
	"reflect"
)

func provide(reg Registry, typ reflect.Type, flaggerBuilder any, buildFunc func(Context) (any, error)) Constructor {
	sf := newStructFlagger(flaggerBuilder)
	if group, ok := reg.constructors[typ]; ok {
		if group.exists(reg.name) {
			if !reg.override {
				panic(fmt.Errorf("类型: %s, 名称: %s已存在", typ, reg.name))
			}
		}
	} else {
		reg.constructors[typ] = newConstructorGroup()
	}
	c := &constructor{
		builder:           flaggerBuilder,
		addFlags:          sf.AddFlags,
		validateFlagsFunc: sf.ValidateFlags,
		buildFunc:         buildFunc,
	}
	if err := reg.constructors[typ].add(reg.name, c); err != nil {
		panic(fmt.Errorf("类型: %s, 名称: %s添加失败: %s", typ, reg.name, err))
	}
	return c
}

func Provide[T any](reg Registry, b Builder[T]) Constructor {
	typ := reflectType[T]()
	return provide(reg, typ, b, func(ctx Context) (any, error) {
		return b.Build(ctx)
	})
}
