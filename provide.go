package di

import (
	"context"
	"flag"
	"fmt"
	"reflect"
)

type options struct {
	name       string
	override   bool
	flagset    *flag.FlagSet
	selections map[reflect.Type]string
	implements map[reflect.Type]reflect.Type
}

func resolveOptions(opts ...Option) options {
	var provideOptions options
	for _, opt := range opts {
		opt.apply(&provideOptions)
	}
	return provideOptions
}

type Option interface {
	apply(*options)
}

type optionFunc func(*options)

func (o optionFunc) apply(opts *options) {
	o(opts)
}

func WithName(name string) Option {
	return optionFunc(func(opts *options) {
		opts.name = name
	})
}

func WithOverride() Option {
	return optionFunc(func(opts *options) {
		opts.override = true
	})
}

func WithFlagset(fs *flag.FlagSet) Option {
	return optionFunc(func(opts *options) {
		opts.flagset = fs
	})
}

func WithSelect[T any](name string) Option {
	return optionFunc(func(opts *options) {
		if opts.selections == nil {
			opts.selections = make(map[reflect.Type]string)
		}
		opts.selections[reflectType[T]()] = name
	})
}

// WithImplement use to specify the implementation of interface
// the first type is interface, the second type is implementation
func WithImplement[I any, T any]() Option {
	return optionFunc(func(opts *options) {
		if opts.implements == nil {
			opts.implements = make(map[reflect.Type]reflect.Type)
		}
		iType := reflectType[I]()
		tType := reflectType[T]()
		if iType.Kind() != reflect.Interface {
			panic(fmt.Errorf("type: %s is not interface", iType))
		}
		if !tType.Implements(iType) {
			panic(fmt.Errorf("type: %s does not implement interface: %s", tType, iType))
		}
		if _, ok := opts.implements[iType]; ok {
			panic(fmt.Errorf("interface: %s already has implementation: %s", iType, opts.implements[iType]))
		}
		opts.implements[iType] = tType
	})
}

func provide(typ reflect.Type, flaggerBuilder any, buildFunc func(context.Context) (any, error), opts ...Option) {
	if typ.Kind() != reflect.Slice && typ.Kind() != reflect.Map {
		// Don't allow providing slices or maps, because they're used to
		// get all instances of the same type.
		// You can provide a type multiple times and use a slice or map to get them all.
		// You can also nest slices or maps inside structs and provide the struct.
		panic(fmt.Errorf("type: %s is not allowed to be provided", typ))
	}

	provideOptions := resolveOptions(opts...)

	sf := newStructFlagger(flaggerBuilder)
	if group, ok := reg.constructors[typ]; ok {
		if group.exists(provideOptions.name) {
			if !provideOptions.override {
				panic(fmt.Errorf("type: %s, Name: %s already exists", typ, provideOptions.name))
			}
		}
	} else {
		reg.constructors[typ] = newConstructorGroup()
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
	if err := reg.constructors[typ].add(provideOptions.name, c); err != nil {
		panic(fmt.Errorf("type: %s, Name: %s add failed: %s", typ, provideOptions.name, err))
	}
}

func Provide[T any](b Builder[T], opts ...Option) {
	provide(reflectType[T](), b, func(ctx context.Context) (any, error) {
		return b.Build(ctx)
	}, opts...)
}
