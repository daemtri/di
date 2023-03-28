package container

import (
	"context"
	"reflect"
)

var (
	ContextKey = &struct{ name string }{name: "di.container.ContextKey"}
)

// ALL 定义了一个通配符, 用于获取所有的依赖
// 用法: Invoke[All[MyInterface]](ctx)
type All[T any] map[string]T

// Container 定义了一个容器, 用于获取对象
type Interface interface {
	// Invoke 获取一个对象, 如果不存在, 则 panic
	Invoke(ctx context.Context, typ reflect.Type) any
}

// Invoke 获取一个对象, 如果不存在, 则 panic
func Invoke[T any](ctx context.Context) T {
	return ctx.Value(ContextKey).(Interface).Invoke(ctx, reflect.TypeOf(new(T)).Elem()).(T)
}

// Simple 定义了一个简单的容器, 可用于mock测试场景
type Simple map[reflect.Type]any

// Put 添加一个对象
func (s Simple) Put(obj any) {
	s[reflect.TypeOf(obj)] = obj
}

// Invoke 获取一个对象, 如果不存在, 则 panic
func (s Simple) Invoke(ctx context.Context, typ reflect.Type) any {
	if v, ok := s[typ]; ok {
		return v
	}
	panic("not found")
}
