package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/handler"
	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/mocks"
	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestMetricOperation_Counter(t *testing.T) {

	log := logger()
	domain.SetMainLogger(log)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testValue := int64(132)
	testValueStr := fmt.Sprintf("%v", testValue)

	m := mocks.NewMockMetricApp(ctrl)

	counterName := "TestCounter"

	m.EXPECT().Get(gomock.Any(), gomock.Eq(domain.CounterType), counterName).Return(&domain.Metrics{
		ID:    counterName,
		MType: domain.CounterType,
		Delta: domain.DeltaPtr(testValue),
	}, nil)

	m.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil, nil)

	r := chi.NewRouter()

	handler.AddMetricOperations(r, m)

	srv := httptest.NewServer(r)
	defer srv.Close()

	req := resty.New().R()
	req.Method = http.MethodPost

	req.URL = srv.URL + "/update/counter/TestCounter/" + testValueStr
	req.Header.Add("Content-Type", handler.TextPlain)
	_, err := req.Send()
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

func TestMetricOperation_Gague(t *testing.T) {

	log := logger()
	domain.SetMainLogger(log)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testValue := float64(132.123)
	testValueStr := fmt.Sprintf("%v", testValue)

	m := mocks.NewMockMetricApp(ctrl)

	gaugeName := "TestGauge"

	m.EXPECT().Get(gomock.Any(), gomock.Eq(domain.GaugeType), gaugeName).Return(&domain.Metrics{
		ID:    gaugeName,
		MType: domain.GaugeType,
		Value: domain.ValuePtr(testValue),
	}, nil)

	m.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil, nil)

	r := chi.NewRouter()

	handler.AddMetricOperations(r, m)

	srv := httptest.NewServer(r)
	defer srv.Close()

	req := resty.New().R()
	req.Method = http.MethodPost

	req.URL = srv.URL + "/update/gauge/TestGauge/" + testValueStr
	req.Header.Add("Content-Type", handler.TextPlain)
	_, err := req.Send()
	require.Nil(t, err)

	req = resty.New().R()
	req.Method = http.MethodGet
	req.URL = srv.URL + "/value/gauge/TestGauge"
	req.Header.Add("Content-Type", handler.TextPlain)

	resp, err := req.Send()
	require.Nil(t, err)
	respBody := string(resp.Body())
	require.Equal(t, testValueStr, respBody)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
}

func TestMetricOperation_All(t *testing.T) {

	log := logger()
	domain.SetMainLogger(log)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockMetricApp(ctrl)

	m.EXPECT().GetAllMetrics(gomock.Any()).Return([]domain.Metrics{
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
	}, nil)

	r := chi.NewRouter()

	handler.AddMetricOperations(r, m)

	srv := httptest.NewServer(r)
	defer srv.Close()

	req := resty.New().R()
	req.Method = http.MethodGet

	req.URL = srv.URL

	resp, err := req.Send()
	require.Nil(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode())
}

func TestMetricOperation_Counter_Update(t *testing.T) {
	log := logger()
	domain.SetMainLogger(log)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockMetricApp(ctrl)

	testValue := int64(2)
	counterName := "PoolCount"

	m.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, m *domain.Metrics) (*domain.Metrics, error) {
			require.NotNil(t, m)
			require.Equal(t, m.MType, domain.CounterType)
			require.NotNil(t, m.Delta)
			require.Equal(t, testValue, *m.Delta)
			return m, nil
		}).Times(1)

	m.EXPECT().Get(gomock.Any(), gomock.Eq(domain.CounterType), gomock.Eq(counterName)).DoAndReturn(
		func(ctx context.Context, metricType domain.MetricType, name string) (*domain.Metrics, error) {
			return &domain.Metrics{
				ID:    counterName,
				MType: domain.CounterType,
				Delta: domain.DeltaPtr(testValue),
			}, nil
		}).Times(1)

	r := chi.NewRouter()

	handler.AddMetricOperations(r, m)

	srv := httptest.NewServer(r)
	defer srv.Close()

	req := resty.New().R()
	req.Method = http.MethodPost

	req.URL = srv.URL + "/update/"
	req.Header.Add("Content-Type", handler.ApplicationJSON)
	metricsReq := domain.Metrics{
		ID:    counterName,
		MType: domain.CounterType,
		Delta: domain.DeltaPtr(testValue),
	}

	req.SetBody(metricsReq)
	_, err := req.Send()
	require.Nil(t, err)

	req.URL = srv.URL + "/value/"
	req.Header.Add("Content-Type", handler.ApplicationJSON)
	metrics := domain.Metrics{
		ID:    counterName,
		MType: domain.CounterType,
	}

	req.SetBody(metrics)
	resp, err := req.Send()
	assert.Nil(t, err)

	var respMetrics domain.Metrics
	err = json.Unmarshal(resp.Body(), &respMetrics)
	require.Nil(t, err)

	require.Equal(t, metricsReq.ID, respMetrics.ID)
	require.Equal(t, metricsReq.MType, respMetrics.MType)
	require.NotNil(t, respMetrics.Delta)
	require.Nil(t, respMetrics.Value)
	require.Equal(t, *metricsReq.Delta, *respMetrics.Delta)
}

func TestMetricOperation_Gague_Update(t *testing.T) {
	log := logger()
	domain.SetMainLogger(log)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockMetricApp(ctrl)

	testValue := float64(123.123)
	gaugeName := "RandomValue"

	m.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, m *domain.Metrics) (*domain.Metrics, error) {
			require.NotNil(t, m)
			require.Equal(t, m.MType, domain.GaugeType)
			require.Nil(t, m.Delta)
			require.NotNil(t, m.Value)
			require.Equal(t, testValue, *m.Value)
			return m, nil
		}).Times(1)

	m.EXPECT().Get(gomock.Any(), gomock.Eq(domain.GaugeType), gomock.Eq(gaugeName)).DoAndReturn(
		func(ctx context.Context, metricType domain.MetricType, name string) (*domain.Metrics, error) {
			require.Equal(t, gaugeName, name)
			return &domain.Metrics{
				ID:    gaugeName,
				MType: domain.GaugeType,
				Value: domain.ValuePtr(testValue),
			}, nil
		}).Times(1)

	r := chi.NewRouter()

	handler.AddMetricOperations(r, m)

	srv := httptest.NewServer(r)
	defer srv.Close()

	req := resty.New().R()
	req.Method = http.MethodPost

	req.URL = srv.URL + "/update/"
	req.Header.Add("Content-Type", handler.ApplicationJSON)
	metricsReq := domain.Metrics{
		ID:    gaugeName,
		MType: domain.GaugeType,
		Value: domain.ValuePtr(testValue),
	}

	req.SetBody(metricsReq)
	_, err := req.Send()
	require.Nil(t, err)

	req.URL = srv.URL + "/value/"
	req.Header.Add("Content-Type", handler.ApplicationJSON)
	metrics := domain.Metrics{
		ID:    gaugeName,
		MType: domain.GaugeType,
	}

	req.SetBody(metrics)
	resp, err := req.Send()
	assert.Nil(t, err)

	var respMetrics domain.Metrics
	err = json.Unmarshal(resp.Body(), &respMetrics)
	require.Nil(t, err)

	require.Equal(t, metricsReq.ID, respMetrics.ID)
	require.Equal(t, metricsReq.MType, respMetrics.MType)
	require.Nil(t, respMetrics.Delta)
	require.NotNil(t, respMetrics.Value)
	require.Equal(t, *metricsReq.Value, *respMetrics.Value)
}

func TestMetricOperation_PostMetrics(t *testing.T) {
	log := logger()
	domain.SetMainLogger(log)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockMetricApp(ctrl)

	testValue := float64(123.123)
	gaugeName := "RandomValue"

	m.EXPECT().UpdateAll(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, ms []domain.Metrics) error {
			require.NotNil(t, ms)
			require.Equal(t, 1, len(ms))
			m := ms[0]
			require.Nil(t, m.Delta)
			require.NotNil(t, m.Value)
			require.Equal(t, testValue, *m.Value)
			return nil
		}).Times(1)

	r := chi.NewRouter()

	handler.AddMetricOperations(r, m)

	srv := httptest.NewServer(r)
	defer srv.Close()

	req := resty.New().R()
	req.Method = http.MethodPost

	req.URL = srv.URL + "/updates/"
	req.Header.Add("Content-Type", handler.ApplicationJSON)
	metricsReq := []domain.Metrics{
		{
			ID:    gaugeName,
			MType: domain.GaugeType,
			Value: domain.ValuePtr(testValue),
		},
	}

	req.SetBody(metricsReq)
	_, err := req.Send()
	require.Nil(t, err)
}

func logger() *zap.SugaredLogger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		// вызываем панику, если ошибка
		panic("cannot initialize zap")
	}
	defer logger.Sync()

	log := logger.Sugar()
	return log
}
