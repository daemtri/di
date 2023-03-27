package object

import (
	"context"
	"reflect"
)

var (
	ContextKey = &contextKey{name: "object container"}
)

type contextKey struct {
	name string
}

type Container interface {
	Exists(ctx context.Context, typ reflect.Type) bool
	Must(ctx context.Context, typ reflect.Type) any
	MustAll(ctx context.Context, typ reflect.Type) map[string]any
}

func Exists[T any](ctx context.Context) bool {
	return ctx.Value(ContextKey).(Container).Exists(ctx, reflect.TypeOf(new(T)).Elem())
}

func Must[T any](ctx context.Context) (t T) {
	return ctx.Value(ContextKey).(Container).Must(ctx, reflect.TypeOf(new(T)).Elem()).(T)
}

func MustAll[T any](ctx context.Context) map[string]T {
	v := ctx.Value(ContextKey).(Container).MustAll(ctx, reflect.TypeOf(new(T)).Elem())
	ret := make(map[string]T, len(v))
	for name := range v {
		ret[name] = v[name].(T)
	}
	return ret
}
