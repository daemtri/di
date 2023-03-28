package di

import (
	"context"
	"fmt"
	"reflect"

	"github.com/daemtri/di/object"
)

func withContext(ctx context.Context, c Context) context.Context {
	return context.WithValue(ctx, object.ContextKey, c)
}

func getContext(ctx context.Context) Context {
	return ctx.Value(object.ContextKey).(Context)
}

// requirer 定义了一个依赖
type requirer struct {
	typ         reflect.Type
	name        string
	constructor *constructor
	parent      *requirer
}

// Context 定义了构建上下文, 用于构建对象, 以及获取依赖, 以及获取上下文
type Context interface {
	Path() string

	container() *container
	requirer() *requirer
	isDiscard() bool
}

// baseContext 定义了基础上下文
type baseContext struct {
	c *container
}

func newBaseContext(c *container) *baseContext {
	return &baseContext{c: c}
}

func (bc *baseContext) container() *container {
	return bc.c
}

func (bc *baseContext) requirer() *requirer {
	return nil
}

func (bc *baseContext) isDiscard() bool {
	return false
}

func (bc *baseContext) Path() string {
	return "@root"
}

type requirerContext struct {
	Context // parent
	r       *requirer
	discard bool
}

func withRequirer(parent Context, r *requirer) *requirerContext {
	return &requirerContext{
		Context: parent,
		r:       r,
	}
}

func (rc *requirerContext) requirer() *requirer {
	return rc.r
}

func (rc *requirerContext) Path() string {
	prefix := rc.Context.Path()
	return prefix + "-->" + fmt.Sprintf("%s(%s)", rc.r.typ, rc.r.name)
}

func (rc *requirerContext) isDiscard() bool {
	return rc.discard
}

func (rc *requirerContext) Invoke(ctx context.Context, typ reflect.Type) any {
	if typ.Kind() == reflect.Map {
		elemTyp := typ.Elem()
		allValues := rc.container().mustAll(ctx, elemTyp)
		all := reflect.MakeMap(typ)
		for name := range allValues {
			all.SetMapIndex(reflect.ValueOf(name), reflect.ValueOf(allValues[name]))
		}
		return all.Interface()
	}
	return rc.container().must(ctx, typ)
}

// checkContext 判断一个构建类型是否已存在
func checkContext(ctx Context) error {
	if contextIsConflict(ctx) {
		return fmt.Errorf("dependency conflicts: %s", ctx.Path())
	}
	return nil
}

func contextIsConflict(ctx Context) bool {
	current := ctx.requirer()
	typ, name := current.typ, current.name
	for {
		current = current.parent
		if current == nil {
			break
		}
		if name == current.name && typ == current.typ {
			return true
		}
	}
	return false
}

func getTypeNameFromContext(ctx context.Context, typ reflect.Type) string {
	secs := getContext(ctx).requirer().constructor.selections
	if secs == nil {
		return ""
	}
	return secs[typ]
}
