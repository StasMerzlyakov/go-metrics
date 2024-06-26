// Package memory contains go-metrics memory storage implementation
package memory

import (
	"context"
	"fmt"

	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
)

func NewStorage() *storage {
	return &storage{
		counterStorage: make(map[string]int64),
		gaugeStorage:   make(map[string]float64),
	}
}

type storage struct {
	counterStorage map[string]int64
	gaugeStorage   map[string]float64
}

func (st *storage) SetAllMetrics(ctx context.Context, in []domain.Metrics) error {
	newCounterStorage := make(map[string]int64)
	newGaugeStorage := make(map[string]float64)

	for _, m := range in {
		switch m.MType {
		case domain.CounterType:
			delta := *m.Delta
			newCounterStorage[m.ID] = delta
		case domain.GaugeType:
			value := *m.Value
			newGaugeStorage[m.ID] = value
		default:
			return fmt.Errorf("unknown MType %v", m.MType)
		}
	}

	st.counterStorage = newCounterStorage
	st.gaugeStorage = newGaugeStorage
	return nil
}

func (st *storage) GetAllMetrics(ctx context.Context) ([]domain.Metrics, error) {

	var out []domain.Metrics

	for k, v := range st.counterStorage {
		delta := v
		out = append(out, domain.Metrics{
			ID:    k,
			MType: domain.CounterType,
			Delta: &delta,
		})
	}

	for k, v := range st.gaugeStorage {
		value := v
		out = append(out, domain.Metrics{
			ID:    k,
			MType: domain.GaugeType,
			Value: &value,
		})
	}

	return out, nil
}

func (st *storage) Set(ctx context.Context, m *domain.Metrics) error {
	switch m.MType {
	case domain.CounterType:
		delta := *m.Delta
		st.counterStorage[m.ID] = delta
	case domain.GaugeType:
		value := *m.Value
		st.gaugeStorage[m.ID] = value
	default:
		return fmt.Errorf("unknown MType %v", m.MType)
	}
	return nil
}

func (st *storage) Add(ctx context.Context, m *domain.Metrics) error {
	switch m.MType {
	case domain.CounterType:
		delta := *m.Delta
		curValue, ok := st.counterStorage[m.ID]
		if ok {
			delta += curValue
			st.counterStorage[m.ID] = delta
			// обновляем значение для входной переменной
			m.Delta = &delta
		} else {
			st.counterStorage[m.ID] = delta
		}
	case domain.GaugeType:
		value := *m.Value
		st.gaugeStorage[m.ID] = value
		curValue, ok := st.gaugeStorage[m.ID]
		if ok {
			curValue += value
			st.gaugeStorage[m.ID] = curValue
			// обновляем значение для входной переменной
			m.Value = &curValue
		} else {
			st.gaugeStorage[m.ID] = value
		}
	default:
		return fmt.Errorf("unknown MType %v", m.MType)
	}
	return nil
}
func (st *storage) Get(ctx context.Context, id string, mType domain.MetricType) (*domain.Metrics, error) {
	switch mType {
	case domain.CounterType:
		curValue, ok := st.counterStorage[id]
		delta := curValue
		if ok {
			return &domain.Metrics{
				ID:    id,
				MType: mType,
				Delta: &delta,
			}, nil
		} else {
			return nil, nil
		}
	case domain.GaugeType:
		curValue, ok := st.gaugeStorage[id]
		value := curValue
		if ok {
			return &domain.Metrics{
				ID:    id,
				MType: mType,
				Value: &value,
			}, nil
		} else {
			return nil, nil
		}
	default:
		return nil, fmt.Errorf("unknown MType %v", mType)
	}
}

func (st *storage) Bootstrap(ctx context.Context) error {
	return nil
}

func (st *storage) Ping(ctx context.Context) error {
	return nil
}

func (st *storage) SetMetrics(ctx context.Context, metric []domain.Metrics) error {
	for _, m := range metric {
		switch m.MType {
		case domain.GaugeType:
			st.gaugeStorage[m.ID] = *m.Value

		case domain.CounterType:
			st.counterStorage[m.ID] = *m.Delta
		}
	}
	return nil
}

func (st *storage) AddMetrics(ctx context.Context, metric []domain.Metrics) error {
	for _, m := range metric {
		switch m.MType {
		case domain.GaugeType:
			value := *m.Value
			curValue, ok := st.gaugeStorage[m.ID]
			if ok {
				curValue += value
				st.gaugeStorage[m.ID] = curValue
			} else {
				st.gaugeStorage[m.ID] = value
			}
		case domain.CounterType:
			delta := *m.Delta
			curValue, ok := st.counterStorage[m.ID]
			if ok {
				delta += curValue
				st.counterStorage[m.ID] = delta
			} else {
				st.counterStorage[m.ID] = delta
			}
		}
	}
	return nil
}

func (st *storage) Close(ctx context.Context) error {
	return nil
}
