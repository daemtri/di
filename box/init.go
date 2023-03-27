package box

import (
	"context"
	"fmt"
)

type initFunc func(ctx context.Context) error

type initializer[T any] struct {
	beforeFunc initFunc
	instance   T
}

func (i *initializer[T]) Build(ctx context.Context) (*initializer[T], error) {
	if err := i.beforeFunc(ctx); err != nil {
		return nil, fmt.Errorf("执行初始化程序出错: %w", err)
	}
	i.instance = Must[T](ctx)
	return i, nil
}
