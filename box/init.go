package box

import (
	"fmt"

	"github.com/daemtri/di"
)

type initFunc func(ctx Context) error

type initializer[T any] struct {
	beforeFunc initFunc
	instance   T
}

func (i *initializer[T]) Build(ctx Context) (*initializer[T], error) {
	if err := i.beforeFunc(ctx); err != nil {
		return nil, fmt.Errorf("执行初始化程序出错: %w", err)
	}
	i.instance = di.Must[T](ctx)
	return i, nil
}
