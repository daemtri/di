package di

import (
	"flag"
	"sync"
)

type Constructor interface {
	validateFlags() error
	build(ctx Context) (any, error)

	AddFlags(fs *flag.FlagSet) Constructor
}

type constructor struct {
	builder  any
	instance any

	addFlags          func(fs *flag.FlagSet)
	validateFlagsFunc func() error
	buildFunc         func(ctx Context) (any, error)

	mux sync.RWMutex
}

func (c *constructor) validateFlags() error {
	return c.validateFlagsFunc()
}

func (c *constructor) build(ctx Context) (any, error) {
	if c.instance != nil {
		return c.instance, nil
	}
	if c.mux.TryLock() {
		defer c.mux.Unlock()
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
