package di

import (
	"context"
	"flag"
	"reflect"
	"sync"
)

type Selection struct {
	typ  reflect.Type
	name string
}

// Constructor 定义了一个构建器
type Constructor interface {
	validateFlags() error
	build(ctx context.Context) (any, error)

	// Designate 为构建器指定一个或多个选择器
	Designate(...Selection) Constructor
	// AddFlags 为构建器添加命令行参数
	AddFlags(fs *flag.FlagSet) Constructor
}

// Select 选择一个构建器
func Select[T any](name string) Selection {
	return Selection{typ: reflectType[T](), name: name}
}

type constructor struct {
	builder  any
	instance any

	addFlags          func(fs *flag.FlagSet)
	validateFlagsFunc func() error
	buildFunc         func(ctx context.Context) (any, error)

	selections map[reflect.Type]string

	mux sync.RWMutex
}

func (c *constructor) validateFlags() error {
	return c.validateFlagsFunc()
}

func (c *constructor) build(ctx context.Context) (any, error) {
	if c.instance != nil {
		return c.instance, nil
	}
	if c.mux.TryLock() {
		defer c.mux.Unlock()
		if err := getContext(ctx).container().inject(ctx, c); err != nil {
			return nil, err
		}
		result, err := c.buildFunc(ctx)
		if err != nil {
			return nil, err
		}
		c.instance = result
	} else {
		c.mux.RLock()
		defer c.mux.RUnlock()
	}
	return c.instance, nil
}

func (c *constructor) AddFlags(fs *flag.FlagSet) Constructor {
	c.addFlags(fs)
	return c
}

func (c *constructor) Designate(selections ...Selection) Constructor {
	if c.selections == nil {
		c.selections = make(map[reflect.Type]string, len(selections))
	}
	for i := range selections {
		c.selections[selections[i].typ] = selections[i].name
	}
	return c
}
