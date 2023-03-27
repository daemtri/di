package di

import (
	"context"
	"fmt"
	"reflect"
)

type contextKey struct {
	name string
}

var (
	ctxKey = &contextKey{name: "di"}
)

func withContext(ctx context.Context, c Context) context.Context {
	return context.WithValue(ctx, ctxKey, c)
}

func getContext(ctx context.Context) Context {
	return ctx.Value(ctxKey).(Context)
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

func (rc *requirerContext) Exists(ctx context.Context, typ reflect.Type) bool {
	return rc.container().exists(ctx, typ)
}

func (rc *requirerContext) MustAll(ctx context.Context, typ reflect.Type) map[string]any {
	return rc.container().mustAll(ctx, typ)
}

func (rc *requirerContext) Must(ctx context.Context, p reflect.Type) any {
	return rc.container().must(ctx, p)
}

// checkContext 判断一个构建类型是否已存在
func checkContext(ctx Context) error {
	if contextIsConflict(ctx) {
		return fmt.Errorf("依赖冲突: %s", ctx.Path())
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
