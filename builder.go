package di

import (
	"context"
)

// Builder defines the object builder
type Builder[T any] interface {
	Build(ctx context.Context) (T, error)
}

// Build builds a specified object. Build cannot build a named object.
func Build[T any](ctx context.Context) (T, error) {
	if err := reg.ValidateFlags(); err != nil {
		return emptyValue[T](), err
	}
	ctx2 := withContext(ctx, newBaseContext(reg.container))
	typ := reflectType[T]()
	v, err := reg.build(ctx2, typ, "")
	if err != nil {
		return emptyValue[T](), err
	}
	return v.(T), err
}
