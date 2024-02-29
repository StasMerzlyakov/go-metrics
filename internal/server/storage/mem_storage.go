package storage

import (
	"fmt"

	"github.com/StasMerzlyakov/go-metrics/internal/server"
)

type memFloat64 struct {
	memStorage[float64]
}

type memInt64 struct {
	memStorage[int64]
}

func NewMemoryFloat64Storage() *memFloat64 {
	return &memFloat64{
		memStorage[float64]{
			storage: make(map[string]float64),
		},
	}
}

func NewMemoryInt64Storage() *memInt64 {
	return &memInt64{
		memStorage[int64]{
			storage: make(map[string]int64),
		},
	}
}

type memValue interface {
	int64 | float64
}

type memStorage[T memValue] struct {
	storage map[string]T
}

func (ms *memStorage[T]) Load(metrics []server.Metrics) error {
	ns := make(map[string]T)

	for _, it := range metrics {
		if it.MType == server.GaugeType {
			if it.Value == nil {
				return fmt.Errorf("load error: value for key %v is nil", it.ID)
			}
			value := *it.Value
			ns[it.ID] = (T)(value)
		} else {
			if it.Delta == nil {
				return fmt.Errorf("load error: delta for key %v is nil", it.ID)
			}
			value := *it.Value
			ns[it.ID] = (T)(value)
		}
	}
	ms.storage = ns
	return nil
}

func (ms memFloat64) Store() ([]server.Metrics, error) {

}

func (ms memFloat64) Load(metrics []server.Metrics) ([]server.Metrics, error) {

}

func (ms *memStorage[T]) Store() ([]server.Metrics, error) {
	// default method
	return nil, fmt.Errorf("unimplemented !!!")
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
