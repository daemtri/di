package di

type BuildFunc[T any] struct {
	// 函数定义在结构体内部，强制不能自动推导Provide时的类型信息，以便能够查阅注册内容
	fn func(Context) (T, error)
}

func (bf BuildFunc[T]) Build(c Context) (T, error) {
	return bf.fn(c)
}

func Func[T any](fn func(ctx Context) (T, error)) BuildFunc[T] {
	return BuildFunc[T]{fn: fn}
}
