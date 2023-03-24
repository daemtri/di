package di

import (
	"reflect"
)

var (
	reg = NewRegistry()
)

// Registry 持有所有注册的构造器
type Registry struct {
	*container
}

// NewRegistry 创建并初始化对象容器
func NewRegistry() Registry {
	return Registry{
		container: &container{
			constructors: make(map[reflect.Type]map[string]Constructor),
		},
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
		for name := range c {
			fn(Value{constructor: c[name].(*constructor), Name: name})
		}
	}
}
