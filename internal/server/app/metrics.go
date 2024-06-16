package app

import (
	"context"
	"fmt"
	"regexp"

	"github.com/pkg/errors"

	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
)

type metricsUseCase struct {
	storage         Storage
	changeListeners []domain.ChangeListener
}

func NewMetrics(storage Storage) *metricsUseCase {
	return &metricsUseCase{
		storage: storage,
	}
}

var nameRegexp = regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9_]*$")

func (mc *metricsUseCase) Get(ctx context.Context, metricType domain.MetricType, name string) (*domain.Metrics, error) {
	switch metricType {
	case domain.CounterType:
		return mc.GetCounter(ctx, name)
	case domain.GaugeType:
		return mc.GetGauge(ctx, name)
	default:
		return nil, fmt.Errorf("%w: unknown metricType '%v'", domain.ErrDataFormat, metricType)
	}
}

func (mc *metricsUseCase) Update(ctx context.Context, mtr *domain.Metrics) (*domain.Metrics, error) {
	if mtr == nil {
		return nil, fmt.Errorf("%w: input is null", domain.ErrDataFormat)
	}

	switch mtr.MType {
	case domain.CounterType:
		if err := mc.AddCounter(ctx, mtr); err != nil {
			return nil, err
		}
		return mtr, nil
	case domain.GaugeType:
		if err := mc.SetGauge(ctx, mtr); err != nil {
			return nil, err
		}
		return mtr, nil
	default:
		return nil, fmt.Errorf("%w: unknown metricType '%v'", domain.ErrDataFormat, mtr.MType)
	}
}

func (mc *metricsUseCase) AddListener(changeListener domain.ChangeListener) {
	mc.changeListeners = append(mc.changeListeners, changeListener)
}

func (mc *metricsUseCase) CheckName(name string) bool {
	return nameRegexp.MatchString(name)
}

func (mc *metricsUseCase) CheckMetrics(m *domain.Metrics) error {
	if !mc.CheckName(m.ID) {
		return errors.Wrap(domain.ErrDataFormat, fmt.Sprintf("wrong metric ID %v", m.ID))
	}

	if m.MType != domain.CounterType && m.MType != domain.GaugeType {
		return errors.Wrap(domain.ErrDataFormat, fmt.Sprintf("metric ID: %v have type %v", m.ID, m.MType))
	}

	if m.MType == domain.CounterType {
		if m.Delta == nil {
			return errors.Wrap(domain.ErrDataFormat, fmt.Sprintf("metric ID: %v have MType %v, but delta is null", m.ID, m.MType))
		}

		if m.Value != nil {
			return errors.Wrap(domain.ErrDataFormat, fmt.Sprintf("metric ID: %v have MType %v, but value is not null", m.ID, m.MType))
		}
	}

	if m.MType == domain.GaugeType {
		if m.Value == nil {
			return errors.Wrap(domain.ErrDataFormat, fmt.Sprintf("metric ID: %v have MType %v, but value is null", m.ID, m.MType))
		}

		if m.Delta != nil {
			return errors.Wrap(domain.ErrDataFormat, fmt.Sprintf("metric ID: %v have MType %v, but delta is not null", m.ID, m.MType))
		}
	}
	return nil
}

func (mc *metricsUseCase) SetAllMetrics(ctx context.Context, in []domain.Metrics) error {
	// Проверка данных
	for _, m := range in {
		err := mc.CheckMetrics(&m)
		if err != nil {
			return err
		}
	}

	return mc.storage.SetAllMetrics(ctx, in)
}

func (mc *metricsUseCase) GetAllMetrics(ctx context.Context) ([]domain.Metrics, error) {
	return mc.storage.GetAllMetrics(ctx)
}

func (mc *metricsUseCase) GetCounter(ctx context.Context, name string) (*domain.Metrics, error) {
	if !mc.CheckName(name) {
		return nil, errors.Wrap(domain.ErrDataFormat, fmt.Sprintf("wrong metric ID %v", name))
	}
	return mc.storage.Get(ctx, name, domain.CounterType)
}

func (mc *metricsUseCase) GetGauge(ctx context.Context, name string) (*domain.Metrics, error) {
	if !mc.CheckName(name) {
		return nil, errors.Wrap(domain.ErrDataFormat, fmt.Sprintf("wrong metric ID %v", name))
	}
	return mc.storage.Get(ctx, name, domain.GaugeType)
}

func (mc *metricsUseCase) AddCounter(ctx context.Context, m *domain.Metrics) error {
	if err := mc.CheckMetrics(m); err != nil {
		return err
	}

	if m.MType != domain.CounterType {
		return fmt.Errorf("unexpected MType %v, expected %v", m.MType, domain.CounterType)
	}

	if err := mc.storage.Add(ctx, m); err != nil {
		return err
	}

	if newValue, err := mc.storage.Get(ctx, m.ID, m.MType); err != nil {
		return err
	} else {
		delta := *newValue.Delta
		m.Delta = &delta
	}

	for _, changeListenerFn := range mc.changeListeners {
		changeListenerFn(ctx, m)
	}

	return nil
}

func (mc *metricsUseCase) SetGauge(ctx context.Context, m *domain.Metrics) error {
	if err := mc.CheckMetrics(m); err != nil {
		return err
	}
	if m.MType != domain.GaugeType {
		return fmt.Errorf("unexpected MType %v, expected %v", m.MType, domain.GaugeType)
	}
	if err := mc.storage.Set(ctx, m); err != nil {
		return err
	}

	for _, changeListenerFn := range mc.changeListeners {
		changeListenerFn(ctx, m)
	}

	return nil
}

func (mc *metricsUseCase) UpdateAll(ctx context.Context, mtr []domain.Metrics) error {
	var gaugeList []domain.Metrics
	var counterList []domain.Metrics

	for _, m := range mtr {
		if err := mc.CheckMetrics(&m); err != nil {
			return err
		}
		switch m.MType {
		case domain.CounterType:
			counterList = append(counterList, m)
		case domain.GaugeType:
			gaugeList = append(gaugeList, m)
		}
	}

	if err := mc.storage.SetMetrics(ctx, gaugeList); err != nil {
		return err
	}

	if err := mc.storage.AddMetrics(ctx, counterList); err != nil {
		return err
	}

	return nil
}
