package di

import (
	"context"
	"fmt"
	"reflect"

	dicontainer "github.com/daemtri/di/container"
)

func withContext(ctx context.Context, c Context) context.Context {
	return context.WithValue(ctx, dicontainer.ContextKey, c)
}

func getContext(ctx context.Context) Context {
	return ctx.Value(dicontainer.ContextKey).(Context)
}

// requirer is defined as a dependency
type requirer struct {
	typ         reflect.Type
	name        string
	constructor *constructor
	parent      *requirer
}

// Context defines the context for building objects, as well as getting dependencies and getting the context itself.
type Context interface {
	Path() string

	container() *container
	requirer() *requirer
	isDiscard() bool
}

// baseContext defines the basic context
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
	} else if typ.Kind() == reflect.Slice {
		elemTyp := typ.Elem()
		allValues := rc.container().mustAll(ctx, elemTyp)
		all := reflect.MakeSlice(typ, 0, len(allValues))
		for _, value := range allValues {
			all = reflect.Append(all, reflect.ValueOf(value))
		}
		return all.Interface()
	}
	return rc.container().must(ctx, typ)
}

// checkContext checks if a build type already exists.
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

func getImplementFromContext(ctx context.Context, typ reflect.Type) reflect.Type {
	imps := getContext(ctx).requirer().constructor.implements
	if imps == nil {
		return nil
	}
	return imps[typ]
}

func getOptionalFuncFromContext(ctx context.Context, typ reflect.Type) func(name string, err error) {
	opts := getContext(ctx).requirer().constructor.optionals
	if opts == nil {
		return nil
	}
	return opts[typ]
}
