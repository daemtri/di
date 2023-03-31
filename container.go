package di

import (
	"context"
	"errors"
	"fmt"
	"reflect"
)

type constructorGroup struct {
	groups map[string]*constructor
}

func newConstructorGroup() *constructorGroup {
	return &constructorGroup{}
}

func (c *constructorGroup) add(name string, cst *constructor) error {
	if c.groups == nil {
		c.groups = make(map[string]*constructor)
	}
	if _, ok := c.groups[name]; ok {
		return fmt.Errorf("builder %s already exists", name)
	}
	c.groups[name] = cst
	return nil
}

func (c *constructorGroup) get(name string) (*constructor, error) {
	if c.groups == nil {
		return nil, fmt.Errorf("builder with name %s does not exist", name)
	}
	rtn, ok := c.groups[name]
	if !ok {
		return nil, fmt.Errorf("builder %s does not exist", name)
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

// validateFlags validates the given flags.
func (c *container) validateFlags() error {
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

func (c *container) build(ctx context.Context, typ reflect.Type, name string) (any, error) {
	localCtx := getContext(ctx)
	if localCtx.isDiscard() {
		return nil, fmt.Errorf("cannot build %s outside of constructor, Context is invalid", typ)
	}
	s, ok := c.constructors[typ]
	if !ok {
		return nil, fmt.Errorf("the type %s (name=[%s]) does not exist在", reflectTypeString(typ), name)
	}
	cst, err := s.get(name)
	if err != nil {
		return nil, fmt.Errorf("type %s (name=[%s]) does not exist: %w", reflectTypeString(typ), name, err)
	}
	newLocalCtx := withRequirer(localCtx, &requirer{
		typ:         typ,
		name:        name,
		constructor: cst,
		parent:      localCtx.requirer(),
	})
	defer func() {
		newLocalCtx.discard = true
	}()
	if err := checkContext(newLocalCtx); err != nil {
		return nil, err
	}
	if err := cst.validateFlags(); err != nil {
		return nil, fmt.Errorf("validate flags error: %w", err)
	}
	rtn, err := cst.build(withContext(ctx, newLocalCtx))
	if err != nil {
		return nil, fmt.Errorf("build type %s error: %w", typ, err)
	}
	return rtn, nil
}

func (c *container) exists(ctx context.Context, p reflect.Type) bool {
	s, ok := c.constructors[p]
	if !ok {
		return false
	}
	localCtx := getContext(ctx)
	return s.exists(localCtx.requirer().constructor.selections[p])
}

func (c *container) mustAll(ctx context.Context, p reflect.Type) map[string]any {
	targetType := p
	if p.Kind() == reflect.Interface {
		if iType := getImplementFromContext(ctx, p); iType != nil {
			targetType = iType
		}
	}
	cst, ok := c.constructors[targetType]
	if !ok {
		panic(fmt.Errorf("the type %s does not exist", reflectTypeString(targetType)))
	}
	vv := make(map[string]any, len(cst.groups))
	optionalFunc := getOptionalFuncFromContext(ctx, p)
	for name := range cst.groups {
		v, err := c.build(ctx, targetType, name)
		if err != nil {
			if optionalFunc != nil {
				optionalFunc(name, err)
				continue
			}
			panic(fmt.Errorf("container build failed: %s", err))
		}
		vv[name] = v
	}

	return vv
}

func (c *container) must(ctx context.Context, p reflect.Type) any {
	targetType := p
	if p.Kind() == reflect.Interface {
		if iType := getImplementFromContext(ctx, p); iType != nil {
			targetType = iType
		}
	}
	name := getTypeNameFromContext(ctx, targetType)
	v, err := c.build(ctx, p, name)
	if err != nil {
		optionalFunc := getOptionalFuncFromContext(ctx, p)
		if optionalFunc != nil {
			optionalFunc(name, err)
			return nil
		}
		panic(fmt.Errorf("container build failed: %s", err))
	}
	return v
}

func (c *container) inject(ctx context.Context, cst *constructor) error {
	refTyp := reflect.TypeOf(cst.builder)
	refVal := reflect.ValueOf(cst.builder)
	if refTyp.Kind() == reflect.Pointer {
		refTyp = refTyp.Elem()
		refVal = refVal.Elem()
	}
	if refTyp.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < refTyp.NumField(); i++ {
		if !refVal.Field(i).CanSet() {
			continue
		}
		injectType, ok := refTyp.Field(i).Tag.Lookup("inject")
		if !ok {
			continue
		}
		if injectType == "must" {
			v := c.must(ctx, refTyp.Field(i).Type)
			refVal.Field(i).Set(reflect.ValueOf(v))
			continue
		}
		if injectType == "exists" {
			if c.exists(ctx, refTyp.Field(i).Type) {
				v := c.must(ctx, refTyp.Field(i).Type)
				refVal.Field(i).Set(reflect.ValueOf(v))
			}
			continue
		}
	}
	return nil
}
