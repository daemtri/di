package object

import (
	"context"
	"reflect"
)

var (
	ContextKey = &struct{ name string }{name: "di.object.ContextKey"}
)

// ALL 定义了一个通配符, 用于获取所有的依赖
type All[T any] map[string]T

// Container 定义了一个容器, 用于获取对象
type Container interface {
	// Invoke 获取一个对象, 如果不存在, 则 panic
	Invoke(ctx context.Context, typ reflect.Type) any
}

// Invoke 获取一个对象, 如果不存在, 则 panic
func Invoke[T any](ctx context.Context) T {
	return ctx.Value(ContextKey).(Container).Invoke(ctx, reflect.TypeOf(new(T)).Elem()).(T)
}

// InvokeAll 获取所有的对象, 如果不存在, 则 panic
func InvokeAll[T any](ctx context.Context) All[T] {
	return ctx.Value(ContextKey).(Container).Invoke(ctx, reflect.TypeOf(new(All[T])).Elem()).(All[T])
}
