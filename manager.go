package di

// Manager 封装了托管对象
type Manager[T any] interface {
	onchangeHandle() error

	// Instance 返回对象实例，当对象不存在时则构建对象
	Instance() T
}
