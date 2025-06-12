package pkg

import (
	"sync"
)

type LockableValue[T any] struct {
	locker sync.RWMutex
	data   T
}

func (l *LockableValue[T]) Get() T {
	l.locker.RLock()
	defer l.locker.RUnlock()
	return l.data
}

func (l *LockableValue[T]) Set(val T) {
	l.locker.Lock()
	defer l.locker.Unlock()
	l.data = val
}

type LockableMap[T any] struct {
	locker sync.RWMutex
	data   map[string]T
	size   int
}

func NewLockableMap[T any](size int) *LockableMap[T] {
	return &LockableMap[T]{
		data: make(map[string]T, size),
		size: size,
	}
}

func (l *LockableMap[T]) Get(key string) (T, bool) {
	l.locker.RLock()
	defer l.locker.RUnlock()
	val, exists := l.data[key]
	return val, exists
}

func (l *LockableMap[T]) Set(key string, val T) {
	l.locker.Lock()
	defer l.locker.Unlock()

	if len(l.data) >= l.size {
		for k := range l.data {
			delete(l.data, k)
			break
		}
	}

	l.data[key] = val
}
