package server

type metricController struct {
	counterStorage CounterStorage
	gaugeStorage   GougeStorage
}

type MemValue interface {
	int64 | float64
}

type MetricsStorage[T MemValue] interface {
	Load(MetricsStorage[T])
	Store() MetricsStorage[T]
	Set(key string, value T)
	Add(key string, value T)
	Get(key string) (T, bool)
	Keys() []string
}

type CounterStorage MetricsStorage[int64]
type GougeStorage MetricsStorage[float64]

func NewMetricController(
	counterStorage CounterStorage,
	gaugeStorage GougeStorage) *metricController {
	return &metricController{
		counterStorage: counterStorage,
		gaugeStorage:   gaugeStorage,
	}
}

func (mc *metricController) GetAllMetrics() []Metrics {
	var items []Metrics
	for _, k := range mc.counterStorage.Keys() {
		v, _ := mc.counterStorage.Get(k)
		items = append(items, Metrics{
			ID:    k,
			MType: CounterType,
			Delta: &v,
		})
	}

	for _, k := range mc.gaugeStorage.Keys() {
		v, _ := mc.gaugeStorage.Get(k)
		items = append(items, Metrics{
			ID:    k,
			MType: GaugeType,
			Value: &v,
		})
	}
	return items
}

func (mc *metricController) GetCounter(name string) (int64, bool) {
	return mc.counterStorage.Get(name)
}

func (mc *metricController) GetGauge(name string) (float64, bool) {
	return mc.gaugeStorage.Get(name)
}

func (mc *metricController) AddCounter(name string, value int64) {
	mc.counterStorage.Add(name, value)
}

func (mc *metricController) SetGauge(name string, value float64) {
	mc.gaugeStorage.Set(name, value)
}
