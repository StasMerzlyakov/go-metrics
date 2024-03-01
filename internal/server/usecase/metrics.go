package usecase

import (
	"fmt"
	"regexp"

	"github.com/pkg/errors"

	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
)

type Storage interface {
	SetAllMetrics(in []domain.Metrics) error
	GetAllMetrics() ([]domain.Metrics, error)
	Set(m *domain.Metrics) error
	Add(m *domain.Metrics) error
	Get(id string, mType domain.MetricType) (*domain.Metrics, error)
}

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

func (mc *metricsUseCase) SetAllMetrics(in []domain.Metrics) error {
	// Проверка данных
	for _, m := range in {
		err := mc.CheckMetrics(&m)
		if err != nil {
			return err
		}
	}

	return mc.storage.SetAllMetrics(in)
}

func (mc *metricsUseCase) GetAllMetrics() ([]domain.Metrics, error) {
	return mc.storage.GetAllMetrics()
}

func (mc *metricsUseCase) GetCounter(name string) (*domain.Metrics, error) {
	if !mc.CheckName(name) {
		return nil, errors.Wrap(domain.ErrDataFormat, fmt.Sprintf("wrong metric ID %v", name))
	}
	return mc.storage.Get(name, domain.CounterType)
}

func (mc *metricsUseCase) GetGauge(name string) (*domain.Metrics, error) {
	if !mc.CheckName(name) {
		return nil, errors.Wrap(domain.ErrDataFormat, fmt.Sprintf("wrong metric ID %v", name))
	}
	return mc.storage.Get(name, domain.GaugeType)
}

func (mc *metricsUseCase) AddCounter(m *domain.Metrics) error {
	if err := mc.CheckMetrics(m); err != nil {
		return err
	}

	if m.MType != domain.CounterType {
		return fmt.Errorf("unexpected MType %v, expected %v", m.MType, domain.CounterType)
	}

	if err := mc.storage.Add(m); err != nil {
		return err
	}

	if newValue, err := mc.storage.Get(m.ID, m.MType); err != nil {
		return err
	} else {
		delta := *newValue.Delta
		m.Delta = &delta
	}

	for _, lst := range mc.changeListeners {
		lst.Refresh(m)
	}

	return nil
}

func (mc *metricsUseCase) SetGauge(m *domain.Metrics) error {
	if err := mc.CheckMetrics(m); err != nil {
		return err
	}
	if m.MType != domain.GaugeType {
		return fmt.Errorf("unexpected MType %v, expected %v", m.MType, domain.GaugeType)
	}
	if err := mc.storage.Set(m); err != nil {
		return err
	}

	for _, lst := range mc.changeListeners {
		lst.Refresh(m)
	}

	return nil
}
