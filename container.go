package di

import (
	"fmt"
	"reflect"
)

type container struct {
	constructors map[reflect.Type]Constructor
}

// ValidateFlags 验证参数
func (c *container) ValidateFlags() error {
	var errs []error
	for i := range c.constructors {
		if err := c.constructors[i].validateFlags(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return multiError(errs)
	}
	return nil
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
	if !ok {
		name := mCtx.Context.name()
		return nil, fmt.Errorf("类型%s(name=%s)不存在", reflectTypeString(typ), name)
	}

	rtn, err := s.build(mCtx)
	if err != nil {
		return nil, fmt.Errorf("构建类型%s出错: %w", typ, err)
	}
	return rtn, nil
}
