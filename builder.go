package di

import (
	"context"
)

// Builder 定义了对象构建器
type Builder[T any] interface {
	Build(ctx context.Context) (T, error)
}

// Build 构建一个指定对象, build不能构建命名对象
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
