package storage

func NewMemoryFloat64Storage() *memStorage[float64] {
	return &memStorage[float64]{
		storage: make(map[string]float64),
	}
}

func NewMemoryInt64Storage() *memStorage[int64] {
	return &memStorage[int64]{
		storage: make(map[string]int64),
	}
}

type memValue interface {
	int64 | float64
}

type memStorage[T memValue] struct {
	storage map[string]T
}

func (ms *memStorage[T]) Keys() []string {
	keys := make([]string, 0, len(ms.storage))
	for k := range ms.storage {
		keys = append(keys, k)
	}
	return keys
}

func (ms *memStorage[T]) Set(key string, value T) {
	ms.storage[key] = value
}

func (ms *memStorage[T]) Add(key string, value T) {
	if curVal, ok := ms.storage[key]; ok {
		ms.storage[key] = curVal + value
	} else {
		ms.storage[key] = value
	}
}

func (ms *memStorage[T]) Get(key string) (T, bool) {
	curVal, ok := ms.storage[key]
	return curVal, ok
}
