package di

import (
	"context"
	"reflect"
	"sync"
)

type Selection struct {
	typ  reflect.Type
	name string
}

type constructor struct {
	builder  any
	instance any

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
