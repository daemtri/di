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
	ctx2 := withContext(ctx, newBaseContext(reg.container))
	typ := reflectType[T]()
	v, err := reg.build(ctx2, typ, "")
	if err != nil {
		return emptyValue[T](), err
	}
	return v.(T), err
}

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
