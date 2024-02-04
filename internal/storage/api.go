package storage

type MemValue interface {
	int64 | float64
}

type MetricsStorage[T MemValue] interface {
	Set(key string, value T)
	Add(key string, value T)
	Get(key string) (T, bool)
	Keys() []string
}
