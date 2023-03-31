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
	optionals  map[reflect.Type]func(name string, err error)
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

func WithOptional[T any](fn func(name string, err error)) Option {
	return optionFunc(func(opts *options) {
		if opts.optionals == nil {
			opts.optionals = make(map[reflect.Type]func(nane string, err error))
		}
		opts.optionals[reflectType[T]()] = fn
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

// Provide is used to provide a type T to the container.
// The provided type T must be a struct or a pointer to a struct, or a interface
func Provide[T any](b Builder[T], opts ...Option) {
	reg.Provide(reflectType[T](), b, func(ctx context.Context) (any, error) {
		return b.Build(ctx)
	}, opts...)
}
