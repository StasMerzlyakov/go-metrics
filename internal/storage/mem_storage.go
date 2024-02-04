package storage

import "sync"

func NewMemoryFloat64Storage() MetricsStorage[float64] {
	return &memStorage[float64]{}
}

func NewMemoryInt64Storage() MetricsStorage[int64] {
	return &memStorage[int64]{}
}

type memStorage[T MemValue] struct {
	mtx     sync.Mutex
	storage map[string]T
}

func (ms *memStorage[T]) Set(key string, value T) {
	ms.mtx.Lock()
	defer ms.mtx.Unlock()
	if ms.storage == nil {
		ms.storage = make(map[string]T)
	}
	ms.storage[key] = value
}

func (ms *memStorage[T]) Add(key string, value T) {
	ms.mtx.Lock()
	defer ms.mtx.Unlock()
	if ms.storage == nil {
		ms.storage = make(map[string]T)
	}
	if curVal, ok := ms.storage[key]; ok {
		ms.storage[key] = curVal + value
	} else {
		ms.storage[key] = value
	}
}

func (ms *memStorage[T]) Get(key string) (T, bool) {
	ms.mtx.Lock()
	defer ms.mtx.Unlock()
	if ms.storage == nil {
		ms.storage = make(map[string]T)
	}
	curVal, ok := ms.storage[key]
	return curVal, ok
}
