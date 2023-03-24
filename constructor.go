package di

import (
	"errors"
	"flag"
	"fmt"
	"sync"
)

type Constructor interface {
	validateFlags() error
	build(ctx Context) (any, error)

	AddFlags(fs *flag.FlagSet) Constructor
	Named(name string) Constructor
	Override() Constructor
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

func (c *constructor) Named(name string) Constructor {
	return c
}

func (c *constructor) Override() Constructor {
	return c
}

type multiConstructor struct {
	cs      map[string]Constructor
	current *constructor
}

func (mc *multiConstructor) exists(name string) bool {
	_, ok := mc.cs[name]
	return ok
}

func (mc *multiConstructor) AddFlags(fs *flag.FlagSet) Constructor {
	mc.current.AddFlags(fs)
	return mc
}

func (mc *multiConstructor) validateFlags() error {
	var err error
	for i := range mc.cs {
		err2 := mc.cs[i].validateFlags()
		if err2 != nil {
			if err == nil {
				err = err2
			} else {
				err = errors.Join(err, err2)
			}
		}
	}
	return err
}

func (mc *multiConstructor) build(ctx Context) (any, error) {
	name := ctx.currentMold().Context.name()
	if name == "" {
		return nil, fmt.Errorf("NamedProvider必须指定名称")
	}
	b, ok := mc.cs[name]
	if !ok {
		return nil, fmt.Errorf("指定name:%s不存在", name)
	}
	ret, err := b.build(ctx)
	if err != nil {
		return nil, err
	}
	return ret, err
}
