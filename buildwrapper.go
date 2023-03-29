package di

import "context"

type BuildFunc[T any] struct {
	// Functions are defined in the structure,
	// forced to not automatically infer the type information provided at registration,
	// so that you can view the registration contents
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
