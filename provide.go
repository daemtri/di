package di

import (
	"fmt"
)

func Provide[T any](reg Registry, b Builder[T]) Constructor {
	typ := reflectType[T]()

	sf := newStructFlagger(b)
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
		builder:           b,
		addFlags:          sf.AddFlags,
		validateFlagsFunc: sf.ValidateFlags,
		buildFunc: func(ctx Context) (any, error) {
			return b.Build(ctx)
		},
	}
	if err := reg.constructors[typ].add(reg.name, c); err != nil {
		panic(fmt.Errorf("类型: %s, 名称: %s添加失败: %s", typ, reg.name, err))
	}
	return c
}
