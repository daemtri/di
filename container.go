package di

import (
	"errors"
	"fmt"
	"reflect"
)

type container struct {
	constructors map[reflect.Type]map[string]Constructor
}

// ValidateFlags 验证参数
func (c *container) ValidateFlags() error {
	var err error
	for i := range c.constructors {
		m := c.constructors[i]
		for name := range m {
			err2 := m[name].validateFlags()
			if err2 != nil {
				if err == nil {
					err = err2
				} else {
					err = errors.Join(err, err2)
				}
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

	// TODO: 判断是否存在
	rtn, err := s[name].build(mCtx)
	if err != nil {
		return nil, fmt.Errorf("构建类型%s出错: %w", typ, err)
	}
	return rtn, nil
}
