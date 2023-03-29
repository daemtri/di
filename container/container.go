package container

import (
	"context"
	"reflect"
)

var (
	// ContextKey is the key used to store the container in the context.
	ContextKey = &struct{ name string }{name: "di.container.ContextKey"}
)

// Set is used to return a set of instances for a particular type.
// This is used to inform the Container that it should return a set of instances
// (as opposed to a single instance) for the specified type.
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
// Simple implements the Container interface, but not implement the Set[T] feature.
// if need Invoke[Set[T]](ctx), please Put Set[T] into Simple.
type Simple map[reflect.Type]any

// SimpleContext create a context with a simple container
// objects can be any type, but must be unique.
// and the type of objects will be used as the key of the container.
// usage:
// ctx := SimpleContext(context.Background(), &MyService{})
//
//	ctx := SimpleContext(context.Background(), &MyService{}, &MyOtherService{})
//
// ctx := SimpleContext(context.Background(), &MyService{}, &MyOtherService{}, Set[MyInterface]{&MyService{}, &MyOtherService{}})
func SimpleContext(ctx context.Context, objects ...any) context.Context {
	s := make(Simple)
	for i := range objects {
		s.Put(objects[i])
	}
	return context.WithValue(ctx, ContextKey, s)
}

// Put an object into the container, the type of the object will be used as the key.
func (s Simple) Put(obj any) {
	if s == nil {
		s = make(Simple)
	}
	s[reflect.TypeOf(obj)] = obj
}

// Invoke Get an object or panic if it doesn't exist
func (s Simple) Invoke(ctx context.Context, typ reflect.Type) any {
	if v, ok := s[typ]; ok {
		return v
	}
	panic("not found")
}
