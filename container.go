package di

import (
	"errors"
	"fmt"
	"reflect"
)

type constructorGroup struct {
	groups map[string]Constructor
}

func newConstructorGroup() *constructorGroup {
	return &constructorGroup{}
}

func (c *constructorGroup) add(name string, constructor Constructor) error {
	if c.groups == nil {
		c.groups = make(map[string]Constructor)
	}
	if _, ok := c.groups[name]; ok {
		return fmt.Errorf("名称为%s的构建器已存在", name)
	}
	c.groups[name] = constructor
	return nil
}

func (c *constructorGroup) get(name string) (Constructor, error) {
	if c.groups == nil {
		return nil, fmt.Errorf("名称为%s的构建器不存在", name)
	}
	rtn, ok := c.groups[name]
	if !ok {
		return nil, fmt.Errorf("名称为%s的构建器不存在", name)
	}
	return rtn, nil
}

func (c *constructorGroup) exists(name string) bool {
	if c.groups == nil {
		return false
	}
	_, ok := c.groups[name]
	return ok
}

func (c *constructorGroup) validateFlags() error {
	var err error
	for i := range c.groups {
		err2 := c.groups[i].validateFlags()
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

type container struct {
	constructors map[reflect.Type]*constructorGroup
}

// ValidateFlags 验证参数
func (c *container) ValidateFlags() error {
	var err error
	for i := range c.constructors {
		m := c.constructors[i]
		err2 := m.validateFlags()
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

func (c *container) build(ctx Context, typ reflect.Type) (any, error) {
	if ctx.isDiscard() {
		return nil, fmt.Errorf("无法在构造函数外构建 %s, Context已失效", typ)
	}
	mCtx := withMold(ctx, typ)
	defer func() {
		mCtx.discard = true
	}()
	if err := checkContext(mCtx); err != nil {
		return nil, err
	}
	s, ok := c.constructors[typ]
	name := mCtx.Context.name()
	if !ok {
		return nil, fmt.Errorf("类型%s(name=%s)不存在", reflectTypeString(typ), name)
	}
	constructor, err := s.get(name)
	if err != nil {
		return nil, fmt.Errorf("类型%s(name=%s)不存在: %w", reflectTypeString(typ), name, err)
	}
	rtn, err := constructor.build(mCtx)
	if err != nil {
		return nil, fmt.Errorf("构建类型%s出错: %w", typ, err)
	}
	return rtn, nil
}
