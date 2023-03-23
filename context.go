package di

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

type Context interface {
	Unwrap() context.Context
	Select(s string) Context
	String() string
	container() *container
	mold() reflect.Type
	name() string
	previousMold() *moldContext
	currentMold() *moldContext
	isDiscard() bool
}

type baseContext struct {
	ctx context.Context
	c   *container
}

func wrapContext(ctx context.Context, c *container) *baseContext {
	return &baseContext{ctx: ctx, c: c}
}

func (bs *baseContext) Select(name string) Context {
	if name == "" {
		panic("Select name不能为空")
	}
	return &nameContext{Context: bs, nameValue: name}
}

func (bs *baseContext) Unwrap() context.Context {
	return bs.ctx
}

func (bs *baseContext) container() *container {
	return bs.c
}

func (bs *baseContext) mold() reflect.Type {
	return nil
}

func (bs *baseContext) name() string {
	return ""
}

func (bs *baseContext) String() string {
	return ""
}

func (bs *baseContext) currentMold() *moldContext {
	return nil
}

func (bs *baseContext) previousMold() *moldContext {
	return nil
}

func (bs *baseContext) isDiscard() bool {
	return false
}

type moldContext struct {
	Context // parent
	typ     reflect.Type

	// discard 用于标识已经构建完成，防止Context被保留指针，在控制流程外进行构建
	discard bool
}

func withMold(parent Context, typ reflect.Type) *moldContext {
	return &moldContext{
		Context: parent,
		typ:     typ,
	}
}

func (mc *moldContext) mold() reflect.Type {
	return mc.typ
}

func (mc *moldContext) Exists(typ reflect.Type) bool {
	_, ok := mc.container().constructors[typ]
	return ok
}

func (mc *moldContext) Must(typ reflect.Type) any {
	v, err := mc.container().build(mc, typ)
	if err != nil {
		panic(err)
	}
	return v
}

func (mc *moldContext) Select(name string) Context {
	if name == "" {
		panic("Select name不能为空")
	}
	return &nameContext{Context: mc, nameValue: name}
}

func (mc *moldContext) currentMold() *moldContext {
	return mc
}

func (mc *moldContext) previousMold() *moldContext {
	return mc.Context.currentMold()
}

func (mc *moldContext) name() string {
	return ""
}

func (mc *moldContext) String() string {
	prefix := mc.Context.String()
	if strings.HasSuffix(prefix, ":") {
		return prefix + mc.typ.String()
	}
	return prefix + "-->" + mc.typ.String()
}

func (mc *moldContext) isDiscard() bool {
	return mc.discard
}

type nameContext struct {
	Context
	nameValue string
}

func (nc *nameContext) Exists(typ reflect.Type) bool {
	b, ok := nc.container().constructors[typ]
	if !ok {
		return false
	}
	return b.(*multiConstructor).exists(nc.nameValue)
}

func (nc *nameContext) Must(typ reflect.Type) any {
	v, err := nc.container().build(nc, typ)
	if err != nil {
		panic(err)
	}
	return v
}

func (nc *nameContext) name() string {
	return nc.nameValue
}

func (nc *nameContext) String() string {
	return nc.Context.String() + "-->" + nc.nameValue + ":"
}

// checkContext 判断一个构建类型是否已存在
func checkContext(ctx Context) error {
	if contextIsConflict(ctx) {
		return fmt.Errorf("依赖冲突: %s", ctx.String())
	}
	return nil
}

func contextIsConflict(ctx Context) bool {
	current := ctx.currentMold()
	typ := current.mold()
	name := current.Context.name()
	for {
		current = current.previousMold()
		if current == nil {
			break
		}

		currentTyp := current.mold()
		if currentTyp == nil {
			break
		}
		currentName := current.Context.name()
		if name == currentName && typ == currentTyp {
			return true
		}
	}
	return false
}
