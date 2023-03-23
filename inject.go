package di

type InjectBuilder[T any, K any] struct {
	Opt K `flag:",nested"`
	fn  func(Context, K) (T, error)
}

func (ib *InjectBuilder[T, K]) Build(c Context) (T, error) {
	return ib.fn(c, ib.Opt)
}

func Inject[T any, K any](fn func(ctx Context, option K) (T, error)) *InjectBuilder[T, K] {
	opt := reflectNew[K]()
	return &InjectBuilder[T, K]{fn: fn, Opt: opt}
}
