package container

import (
	"context"
	"reflect"
)

var (
	ContextKey = &struct{ name string }{name: "di.container.ContextKey"}
)

// Set This container is used to get objects
// usage: Invoke[Set[MyInterface]](ctx)
// Note that this feature depends on the container implementation.
// If the container allows to register multiple objects of the same type,
// then Invoke[Set[MyInterface]](ctx) will return all objects of the same type.
type Set[T any] []T

// Interface Container interface
type Interface interface {
	// Invoke Get a value from the map for a key, or panic if the key does not exist.
	Invoke(ctx context.Context, typ reflect.Type) any
}

// Invoke Get a value from the map for a key, or panic if none exists.
func Invoke[T any](ctx context.Context) T {
	return ctx.Value(ContextKey).(Interface).Invoke(ctx, reflect.TypeOf(new(T)).Elem()).(T)
}

// Simple A simple container for mocking testing scenes.
type Simple map[reflect.Type]any

// Put add an object
func (s Simple) Put(obj any) {
	s[reflect.TypeOf(obj)] = obj
}

// Invoke Get an object or panic if it doesn't exist
func (s Simple) Invoke(ctx context.Context, typ reflect.Type) any {
	if v, ok := s[typ]; ok {
		return v
	}
	panic("not found")
}
