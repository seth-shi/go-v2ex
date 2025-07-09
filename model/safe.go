package model

import (
	"sync"
)

type Safe[T any] struct {
	locker      sync.RWMutex
	data        T
	afterUpdate func(v T) error
}

func NewSafe[T any](v T, fn func(v T) error) *Safe[T] {
	return &Safe[T]{
		data:        v,
		afterUpdate: fn,
	}
}

func (s *Safe[T]) Get() T {
	s.locker.RLock()
	s.locker.RUnlock()

	return s.data
}

func (s *Safe[T]) Set(v T) {
	s.locker.Lock()
	s.locker.Unlock()
	s.data = v
}

func (s *Safe[T]) Update(fn func(T)) {
	s.locker.Lock()
	s.locker.Unlock()
	fn(s.data)
}

func (s *Safe[T]) Save(fn func(T)) error {
	s.locker.Lock()
	s.locker.Unlock()
	fn(s.data)

	if s.afterUpdate == nil {
		return nil
	}

	return s.afterUpdate(s.data)
}
