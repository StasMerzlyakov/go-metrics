package handler_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/StasMerzlyakov/go-metrics/internal/server/controller/http/handler"
	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestExtractFloat64(t *testing.T) {
	type extractFloat64Result struct {
		value             float64
		isSuccessExpected bool
	}
	tests := []struct {
		name   string
		input  string
		result extractFloat64Result
	}{
		{
			"good value",
			"123.5",
			extractFloat64Result{
				123.5,
				true,
			},
		},
		{
			"good value 2",
			"123",
			extractFloat64Result{
				123,
				true,
			},
		},
		{
			"bad value",
			"123.F",
			extractFloat64Result{
				-1,
				false,
			},
		},
		{
			"good value",
			"1.8070544e+07",
			extractFloat64Result{
				18070544,
				true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := handler.ExtractFloat64(tt.input)
			assert.Equal(t, tt.result.value, value)
			assert.Equal(t, tt.result.isSuccessExpected, err == nil)
		})
	}
}

func TestExtractInt64(t *testing.T) {

	type extractInt64Result struct {
		value             int64
		isSuccessExpected bool
	}

	tests := []struct {
		name   string
		input  string
		result extractInt64Result
	}{
		{
			"good value",
			"123",
			extractInt64Result{
				123,
				true,
			},
		},
		{
			"bad value",
			"123F",
			extractInt64Result{
				-1,
				false,
			},
		},
		{
			"bad value 2",
			"123.5",
			extractInt64Result{
				-1,
				false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := handler.ExtractInt64(tt.input)
			assert.Equal(t, tt.result.value, value)
			assert.Equal(t, tt.result.isSuccessExpected, err == nil)
		})
	}
}

type mockMetricUse struct {
	counterVal int64
	gaugeVal   float64
}

func (*mockMetricUse) SetAllMetrics(in []domain.Metrics) error {
	return nil
}
func (*mockMetricUse) GetAllMetrics() ([]domain.Metrics, error) {
	return []domain.Metrics{
		{
			ID:    "PoolCount",
			MType: domain.CounterType,
			Delta: domain.DeltaPtr(1),
		},
		{
			ID:    "RandomValue",
			MType: domain.GaugeType,
			Value: domain.ValuePtr(1.1),
		},
	}, nil
}
func (m *mockMetricUse) GetCounter(name string) (*domain.Metrics, error) {
	return &domain.Metrics{
		ID:    name,
		MType: domain.CounterType,
		Delta: domain.DeltaPtr(m.counterVal),
	}, nil
}
func (m *mockMetricUse) GetGauge(name string) (*domain.Metrics, error) {
	return &domain.Metrics{
		ID:    name,
		MType: domain.GaugeType,
		Value: domain.ValuePtr(m.gaugeVal),
	}, nil
}
func (*mockMetricUse) AddCounter(m *domain.Metrics) error {
	return nil
}
func (*mockMetricUse) SetGauge(m *domain.Metrics) error {
	return nil
}

func TestPostUpdate(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		// вызываем панику, если ошибка
		panic("cannot initialize zap")
	}
	defer logger.Sync()

	log := logger.Sugar()

	hHandler := handler.NewHTTP(&mockMetricUse{}, log)

	srv := httptest.NewServer(hHandler)
	defer srv.Close()

	req := resty.New().R()
	req.Method = http.MethodPost

	req.URL = srv.URL + "/update/"
	req.Header.Add("Content-Type", handler.ApplicationJSON)
	_, err = req.Send()
	require.Nil(t, err)

	req.URL = srv.URL + "/value/"
	req.Header.Add("Content-Type", handler.ApplicationJSON)
	_, err = req.Send()
	assert.Nil(t, err)
}

func TestCounterValueHandler(t *testing.T) {
	testValue := 123
	testValueStr := fmt.Sprintf("%v", testValue)

	logger, err := zap.NewDevelopment()
	if err != nil {
		// вызываем панику, если ошибка
		panic("cannot initialize zap")
	}
	defer logger.Sync()

	log := logger.Sugar()

	hHandler := handler.NewHTTP(&mockMetricUse{
		counterVal: int64(testValue),
	}, log)

	srv := httptest.NewServer(hHandler)
	defer srv.Close()

	req := resty.New().R()
	req.Method = http.MethodPost

	req.URL = srv.URL + "/update/counter/TestCounter/" + testValueStr
	req.Header.Add("Content-Type", handler.TextPlain)
	_, err = req.Send()
	require.Nil(t, err)

	req = resty.New().R()
	req.Method = http.MethodGet
	req.URL = srv.URL + "/value/counter/TestCounter"
	req.Header.Add("Content-Type", handler.TextPlain)

	resp, err := req.Send()
	require.Nil(t, err)
	respBody := string(resp.Body())
	require.Equal(t, testValueStr, respBody)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
}

func TestGaugeValueHandler(t *testing.T) {

	logger, err := zap.NewDevelopment()
	if err != nil {
		// вызываем панику, если ошибка
		panic("cannot initialize zap")
	}
	defer logger.Sync()

	log := logger.Sugar()

	testValue := 234.123
	testValueStr := fmt.Sprintf("%v", testValue)

	hHandler := handler.NewHTTP(&mockMetricUse{
		gaugeVal: float64(testValue),
	}, log)

	srv := httptest.NewServer(hHandler)
	defer srv.Close()

	req := resty.New().R()
	req.Method = http.MethodPost
	req.URL = srv.URL + "/update/gauge/TestCounter/" + testValueStr
	req.Header.Add("Content-Type", handler.TextPlain)
	_, err = req.Send()
	require.Nil(t, err)

	req = resty.New().R()
	req.Method = http.MethodGet
	req.URL = srv.URL + "/value/gauge/TestCounter"
	req.Header.Add("Content-Type", handler.TextPlain)
	resp, err := req.Send()
	require.Nil(t, err)
	respBody := string(resp.Body())
	require.Equal(t, testValueStr, respBody) // Не попасть бы на потерю точности string -> float64 -> string
	assert.Equal(t, http.StatusOK, resp.StatusCode())
}

func TestGetAll(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		// вызываем панику, если ошибка
		panic("cannot initialize zap")
	}
	defer logger.Sync()

	log := logger.Sugar()

	hHandler := handler.NewHTTP(&mockMetricUse{}, log)

	srv := httptest.NewServer(hHandler)
	defer srv.Close()

	req := resty.New().R()
	req.Method = http.MethodGet

	req.URL = srv.URL

	resp, err := req.Send()
	require.Nil(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode())
}

type mockMetricUse2 struct {
	metrics *domain.Metrics
}

func (*mockMetricUse2) SetAllMetrics(in []domain.Metrics) error {
	return nil
}
func (m2 *mockMetricUse2) GetAllMetrics() ([]domain.Metrics, error) {
	var out []domain.Metrics
	if m2.metrics != nil {
		out = append(out, *m2.metrics)
	}
	return out, nil

}
func (m2 *mockMetricUse2) GetCounter(name string) (*domain.Metrics, error) {
	if m2.metrics != nil {
		if m2.metrics.ID == name && m2.metrics.MType == domain.CounterType {
			return m2.metrics, nil
		}
	}
	return nil, nil
}
func (m2 *mockMetricUse2) GetGauge(name string) (*domain.Metrics, error) {
	if m2.metrics != nil {
		if m2.metrics.ID == name && m2.metrics.MType == domain.GaugeType {
			return m2.metrics, nil
		}
	}
	return nil, nil
}
func (m2 *mockMetricUse2) AddCounter(m *domain.Metrics) error {
	if m.MType == domain.CounterType {
		m2.metrics = m
	} else {
		return fmt.Errorf("unexpected input")
	}
	return nil
}
func (m2 *mockMetricUse2) SetGauge(m *domain.Metrics) error {
	if m.MType == domain.GaugeType {
		m2.metrics = m
	} else {
		return fmt.Errorf("unexpected input")
	}
	return nil
}

func TestCounterPostAndValue(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		// вызываем панику, если ошибка
		panic("cannot initialize zap")
	}
	defer logger.Sync()

	log := logger.Sugar()

	hHandler := handler.NewHTTP(&mockMetricUse2{}, log)

	srv := httptest.NewServer(hHandler)
	defer srv.Close()

	// Метрик нет
	req := resty.New().R()
	req.Method = http.MethodGet
	req.URL = srv.URL

	resp, err := req.Send()
	require.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())

	req.Method = http.MethodPost
	req.Header.Add("Content-Type", handler.ApplicationJSON)
	req.URL = srv.URL + "/value"
	metrics := domain.Metrics{
		ID:    "PoolCount",
		MType: domain.CounterType,
	}

	req.SetBody(metrics)

	resp, err = req.Send()
	require.Nil(t, err)
	require.Equal(t, resp.StatusCode(), http.StatusNotFound)

	// Добавляем метрику

	req.Header.Add("Content-Type", handler.ApplicationJSON)
	req.URL = srv.URL + "/update"

	metricsReq := domain.Metrics{
		ID:    "PoolCount",
		MType: domain.CounterType,
		Delta: domain.DeltaPtr(2),
	}

	req.SetBody(metricsReq)

	resp, err = req.Send()
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())

	// Проверяем метрику
	req.Method = http.MethodPost
	req.Header.Add("Content-Type", handler.ApplicationJSON)
	req.URL = srv.URL + "/value"
	metrics = domain.Metrics{
		ID:    metricsReq.ID,
		MType: metricsReq.MType,
	}

	req.SetBody(metrics)

	resp, err = req.Send()
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())

	var respMetrics domain.Metrics
	err = json.Unmarshal(resp.Body(), &respMetrics)
	require.Nil(t, err)

	require.Equal(t, metricsReq.ID, respMetrics.ID)
	require.Equal(t, metricsReq.MType, respMetrics.MType)
	require.NotNil(t, respMetrics.Delta)
	require.Nil(t, respMetrics.Value)
	require.Equal(t, *metricsReq.Delta, *respMetrics.Delta)
}

func TestGaguePostAndValue(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		// вызываем панику, если ошибка
		panic("cannot initialize zap")
	}
	defer logger.Sync()

	log := logger.Sugar()

	hHandler := handler.NewHTTP(&mockMetricUse2{}, log)

	srv := httptest.NewServer(hHandler)
	defer srv.Close()

	// Метрик нет
	req := resty.New().R()
	req.Method = http.MethodGet
	req.URL = srv.URL

	resp, err := req.Send()
	require.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())

	req.Method = http.MethodPost
	req.Header.Add("Content-Type", handler.ApplicationJSON)
	req.URL = srv.URL + "/value"
	metrics := domain.Metrics{
		ID:    "RandomValue",
		MType: domain.GaugeType,
	}

	req.SetBody(metrics)

	resp, err = req.Send()
	require.Nil(t, err)
	require.Equal(t, resp.StatusCode(), http.StatusNotFound)

	// Добавляем метрику

	req.Header.Add("Content-Type", handler.ApplicationJSON)
	req.URL = srv.URL + "/update"

	metricsReq := domain.Metrics{
		ID:    "RandomValue",
		MType: domain.GaugeType,
		Value: domain.ValuePtr(2),
	}

	req.SetBody(metricsReq)

	resp, err = req.Send()
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())

	// Проверяем метрику
	req.Method = http.MethodPost
	req.Header.Add("Content-Type", handler.ApplicationJSON)
	req.URL = srv.URL + "/value"
	metrics = domain.Metrics{
		ID:    metricsReq.ID,
		MType: metricsReq.MType,
	}

	req.SetBody(metrics)

	resp, err = req.Send()
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())

	var respMetrics domain.Metrics
	err = json.Unmarshal(resp.Body(), &respMetrics)
	require.Nil(t, err)

	require.Equal(t, metricsReq.ID, respMetrics.ID)
	require.Equal(t, metricsReq.MType, respMetrics.MType)
	require.Nil(t, respMetrics.Delta)
	require.NotNil(t, respMetrics.Value)
	require.Equal(t, *metricsReq.Value, *respMetrics.Value)
}
