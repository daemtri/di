package box

import (
	"github.com/daemtri/di"
)

type options struct {
	opts []di.Option
}

func newOptions() *options {
	return &options{
		opts: make([]di.Option, 0, 4),
	}
}

type Options interface {
	apply(o *options)
}

type optionsFunc func(o *options)

func (of optionsFunc) apply(o *options) { of(o) }

func WithName(name string) Options {
	return optionsFunc(func(o *options) {
		o.opts = append(o.opts, di.WithName(name))
	})
}

func WithFlagPrefix(prefix string) Options {
	return optionsFunc(func(o *options) {
		o.opts = append(o.opts, di.WithFlagset(nfs.FlagSet(prefix)))
	})
}

// WithSelect 仅供在ProvideInject时使用，可以指定注入某个类型的名字
func WithSelect[T any](name string) Options {
	return optionsFunc(func(o *options) {
		o.opts = append(o.opts, di.WithSelect[T](name))
	})
}

func WithOverride() Options {
	return optionsFunc(func(o *options) {
		o.opts = append(o.opts, di.WithOverride())
	})
}
