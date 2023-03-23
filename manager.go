package di

import (
	"sync"
)

// Manager 封装了托管对象
type Manager[T any] interface {
	onchangeHandle() error

	// Instance 返回对象实例，当对象不存在时则构建对象
	Instance() T
}

// simpleManager 托管对象，当配置发生变更时，由di负责对象的更新等操作
type simpleManager[T any] struct {
	buildFunc func() (T, error)
	mux       sync.RWMutex
	object    T
}

func (sm *simpleManager[T]) onchangeHandle() error {
	newObj, err := sm.buildFunc()
	if err != nil {
		return err
	}
	sm.mux.Lock()
	sm.object = newObj
	sm.mux.Unlock()
	return nil
}

func (sm *simpleManager[T]) Instance() T {
	sm.mux.RLock()
	defer sm.mux.RUnlock()
	return sm.object
}

type managerBuilder[T any] struct {
	FlagSetWatcher

	builder Builder[T]
	manager Manager[T]
}

func (m *managerBuilder[T]) Build(ctx Context) (Manager[T], error) {
	var mr Manager[T]
	x, err := m.builder.Build(ctx)
	if err != nil {
		return nil, err
	}
	mr = &simpleManager[T]{
		object: x,
		buildFunc: func() (T, error) {
			return m.builder.Build(ctx)
		},
	}
	m.manager = mr
	return mr, nil
}

func Manage[T any](b Builder[T]) Builder[Manager[T]] {
	return &managerBuilder[T]{
		builder: b,
	}
}
