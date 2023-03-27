package di

import "context"

type BuildFunc[T any] struct {
	// 函数定义在结构体内部，强制不能自动推导Provide时的类型信息，以便能够查阅注册内容
	fn func(context.Context) (T, error)
}

func (bf BuildFunc[T]) Build(c context.Context) (T, error) {
	return bf.fn(c)
}

func Func[T any](fn func(ctx context.Context) (T, error)) BuildFunc[T] {
	return BuildFunc[T]{fn: fn}
}

type InjectBuilder[T any, K any] struct {
	Opt K `flag:""`
	fn  func(context.Context, K) (T, error)
}

func (ib *InjectBuilder[T, K]) Build(c context.Context) (T, error) {
	return ib.fn(c, ib.Opt)
}

func Inject[T any, K any](fn func(ctx context.Context, option K) (T, error)) *InjectBuilder[T, K] {
	opt := reflectNew[K]()
	return &InjectBuilder[T, K]{fn: fn, Opt: opt}
}
