package di

import (
	"reflect"
)

// Registry 持有所有注册的构造器
type Registry struct {
	*container

	name     string
	override bool
}

// NewRegistry 创建并初始化对象容器
func NewRegistry() Registry {
	return Registry{
		container: &container{
			constructors: make(map[reflect.Type]Constructor),
		},
	}
}

func (r Registry) Override() Registry {
	return Registry{
		name:      r.name,
		override:  true,
		container: r.container,
	}
}

func (r Registry) Named(name string) Registry {
	if name == "" {
		panic("注册命名对象错误: name为空字符串")
	}
	return Registry{
		name:      name,
		override:  r.override,
		container: r.container,
	}
}

type Value struct {
	*constructor
	Name string
}

func (v Value) Instance() any {
	return v.constructor.instance
}

func (v Value) Builder() any {
	return v.constructor.builder
}

// VisitAll 遍历所有已经构建的构造器
func (r Registry) Visit(fn func(v Value)) {
	for _, c := range r.constructors {
		switch v := c.(type) {
		case *constructor:
			fn(Value{constructor: v})
		case *multiConstructor:
			for k, nv := range v.cs {
				fn(Value{Name: k, constructor: nv.(*constructor)})
			}
		}
	}
}
