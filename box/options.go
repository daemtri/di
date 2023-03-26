package box

import (
	"github.com/daemtri/di"
)

type options struct {
	name          string
	flagSetPrefix string

	selects  []di.Selection
	override bool
}

func newOptions() *options {
	return &options{
		name:          "",
		flagSetPrefix: "",
		selects:       nil,
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
		if o.selects == nil {
			o.selects = []di.Selection{di.Select[T](name)}
		} else {
			o.selects = append(o.selects, di.Select[T](name))
		}
	})
}

func WithOverride() Options {
	return optionsFunc(func(o *options) {
		o.override = true
	})
}
