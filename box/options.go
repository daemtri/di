package box

import "reflect"

type options struct {
	name          string
	flagSetPrefix string

	selects  map[reflect.Type]string
	override bool
}

func newOptions() *options {
	return &options{
		name:          "",
		flagSetPrefix: "",
		selects:       make(map[reflect.Type]string),
	}
}

type Options interface {
	apply(o *options)
}

type optionsFunc func(o *options)

func (of optionsFunc) apply(o *options) { of(o) }

func WithName(name string) Options {
	return optionsFunc(func(o *options) {
		o.name = name
	})
}

func WithFlagPrefix(prefix string) Options {
	return optionsFunc(func(o *options) {
		o.flagSetPrefix = prefix
	})
}

// WithSelect 仅供在ProvideInject时使用，可以指定注入某个类型的名字
func WithSelect[T any](name string) Options {
	return optionsFunc(func(o *options) {
		typ := reflect.TypeOf(emptyValue[T]())
		o.selects[typ] = name
	})
}

func WithOverride() Options {
	return optionsFunc(func(o *options) {
		o.override = true
	})
}
