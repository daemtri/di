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

func provide(typ reflect.Type, flaggerBuilder any, buildFunc func(context.Context) (any, error), opts ...Option) {
	provideOptions := resolveOptions(opts...)

	sf := newStructFlagger(flaggerBuilder)
	if group, ok := reg.constructors[typ]; ok {
		if group.exists(provideOptions.name) {
			if !provideOptions.override {
				panic(fmt.Errorf("类型: %s, 名称: %s已存在", typ, provideOptions.name))
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
	}
	if err := reg.constructors[typ].add(provideOptions.name, c); err != nil {
		panic(fmt.Errorf("类型: %s, 名称: %s添加失败: %s", typ, provideOptions.name, err))
	}
}

func Provide[T any](b Builder[T], opts ...Option) {
	provide(reflectType[T](), b, func(ctx context.Context) (any, error) {
		return b.Build(ctx)
	}, opts...)
}
