package di

import (
	"reflect"
)

var (
	reg = NewRegistry()
)

func GetRegistry() Registry {
	return reg
}

// Registry holds all the registered constructors
type Registry struct {
	*container
}

// NewRegistry create and initialize object container
func NewRegistry() Registry {
	return Registry{
		container: &container{
			constructors: make(map[reflect.Type]*constructorGroup),
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

// VisitAll Iterate all the built constructors
func (r Registry) Visit(fn func(v Value)) {
	for _, c := range r.constructors {
		for name, v := range c.groups {
			fn(Value{
				Name:        name,
				constructor: v,
			})
		}
	}
}
