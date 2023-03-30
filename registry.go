package di

import (
	"context"
	"fmt"
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

func (r Registry) Provide(typ reflect.Type, flaggerBuilder any, buildFunc func(context.Context) (any, error), opts ...Option) {
	if typ.Kind() == reflect.Slice || typ.Kind() == reflect.Map {
		// Don't allow providing slices or maps, because they're used to
		// get all instances of the same type.
		// You can provide a type multiple times and use a slice or map to get them all.
		// You can also nest slices or maps inside structs and provide the struct.
		panic(fmt.Errorf("type: %s is not allowed to be provided", typ))
	}

	provideOptions := resolveOptions(opts...)
	sf := newStructFlagger(flaggerBuilder)
	if group, ok := r.constructors[typ]; ok {
		if group.exists(provideOptions.name) {
			if !provideOptions.override {
				panic(fmt.Errorf("type: %s, Name: %s already exists", typ, provideOptions.name))
			}
		}
	} else {
		r.constructors[typ] = newConstructorGroup()
	}
	if provideOptions.flagset != nil {
		sf.AddFlags(provideOptions.flagset)
	}
	c := &constructor{
		builder:           flaggerBuilder,
		validateFlagsFunc: sf.ValidateFlags,
		buildFunc:         buildFunc,
		selections:        provideOptions.selections,
		implements:        provideOptions.implements,
	}
	if err := r.constructors[typ].add(provideOptions.name, c); err != nil {
		panic(fmt.Errorf("type: %s, Name: %s add failed: %s", typ, provideOptions.name, err))
	}
}
