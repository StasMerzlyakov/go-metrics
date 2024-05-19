package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"

	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
	"github.com/go-chi/chi/v5"
	chiMW "github.com/go-chi/chi/v5/middleware"

	_ "github.com/golang/mock/gomock"        // обязательно, требуется в сгенерированных mock-файлах,
	_ "github.com/golang/mock/mockgen/model" // обязательно для корректного запуска mockgen
)

var ErrMediaType = errors.New("UnsupportedMediaTypeError") // ошибку определяю здесь - она специфична для

//go:generate mockgen -destination "../mocks/$GOFILE" -package mocks . MetricApp

type MetricApp interface {
	GetAllMetrics(ctx context.Context) ([]domain.Metrics, error)
	Get(ctx context.Context, metricType domain.MetricType, name string) (*domain.Metrics, error)
	UpdateAll(ctx context.Context, mtr []domain.Metrics) error
	Update(ctx context.Context, mtr *domain.Metrics) (*domain.Metrics, error)
}

func AddMetricOperations(r *chi.Mux, metricApp MetricApp, changeDataMw ...func(http.Handler) http.Handler) {
	adapter := &metricOperationAdapter{
		metricApp: metricApp,
	}

	r.Get("/", adapter.AllMetrics)

	r.Route("/updates", func(r chi.Router) {
		r.Use(changeDataMw...)
		r.Post("/", adapter.PostMetrics)
	})

	r.Route("/update", func(r chi.Router) {
		r.Post("/", adapter.PostMetric)
		r.Post("/gauge/{name}/{value}", adapter.PostGauge)
		r.Post("/gauge/{name}", StatusNotFound)
		r.Post("/counter/{name}/{value}", adapter.PostCounter)
		r.Post("/counter/{name}", StatusNotFound)
		r.Post("/{type}/{name}/{value}", StatusNotImplemented)
	})

	r.Route("/value", func(r chi.Router) {
		r.Post("/", adapter.ValueMetric)
		r.Get("/gauge/{name}", adapter.GetGauge)
		r.Get("/counter/{name}", adapter.GetCounter)
	})

}

func AddPProfOperations(r *chi.Mux) {
	r.Mount("/debug", chiMW.Profiler())
}

type metricOperationAdapter struct {
	metricApp MetricApp
}

func (h *metricOperationAdapter) PostMetrics(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	if err := h.checkContentType(ApplicationJSON, req); err != nil {
		handleAppError(w, err)
		return
	}

	var metrics []domain.Metrics
	if err := json.NewDecoder(req.Body).Decode(&metrics); err != nil {
		fullErr := fmt.Errorf("%w: json decode error - %v", domain.ErrDataFormat, err.Error())
		handleAppError(w, fullErr)
		return
	}

	if err := h.metricApp.UpdateAll(req.Context(), metrics); err != nil {
		handleAppError(w, err)
		return
	}

}

func (h *metricOperationAdapter) PostMetric(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	if err := h.checkContentType(ApplicationJSON, req); err != nil {
		handleAppError(w, err)
		return
	}

	var metrics *domain.Metrics

	if err := json.NewDecoder(req.Body).Decode(&metrics); err != nil {
		fullErr := fmt.Errorf("%w: json decode error - %v", domain.ErrDataFormat, err.Error())
		handleAppError(w, fullErr)
		return
	}

	updatedMetric, err := h.metricApp.Update(req.Context(), metrics)
	if err != nil {
		handleAppError(w, err)
		return
	}

	if err := h.sendMetrics(w, updatedMetric); err != nil {
		handleAppError(w, err)
		return
	}
}

func (h *metricOperationAdapter) ValueMetric(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	if err := h.checkContentType(ApplicationJSON, req); err != nil {
		handleAppError(w, err)
		return
	}

	var metrics *domain.Metrics
	if err := json.NewDecoder(req.Body).Decode(&metrics); err != nil {
		fullErr := fmt.Errorf("%w: json decode error - %v", domain.ErrDataFormat, err.Error())
		handleAppError(w, fullErr)
		return
	}

	value, err := h.metricApp.Get(req.Context(), metrics.MType, metrics.ID)
	if err != nil {
		handleAppError(w, err)
		return
	}

	if value == nil {
		err := fmt.Errorf("%w: unknown metric '%v'", domain.ErrNotFound, metrics.ID)
		handleAppError(w, err)
		return
	}

	if err := h.sendMetrics(w, value); err != nil {
		handleAppError(w, err)
		return
	}
}

func (h *metricOperationAdapter) PostGauge(w http.ResponseWriter, req *http.Request) {
	_, _ = io.ReadAll(req.Body)
	defer req.Body.Close()

	if err := h.checkContentType(TextPlain, req); err != nil {
		handleAppError(w, err)
		return
	}

	var name string
	var value float64
	var err error

	if name, err = h.extractName(req); err != nil {
		handleAppError(w, err)
		return
	}

	if value, err = h.extractFloat64(req); err != nil {
		handleAppError(w, err)
		return
	}

	metrics := &domain.Metrics{
		ID:    name,
		MType: domain.GaugeType,
		Value: domain.ValuePtr(value),
	}

	if _, err := h.metricApp.Update(req.Context(), metrics); err != nil {
		handleAppError(w, err)
		return
	}
}

func (h *metricOperationAdapter) PostCounter(w http.ResponseWriter, req *http.Request) {
	_, _ = io.ReadAll(req.Body)
	defer req.Body.Close()

	if err := h.checkContentType(TextPlain, req); err != nil {
		handleAppError(w, err)
		return
	}

	var name string
	var value int64
	var err error

	if name, err = h.extractName(req); err != nil {
		handleAppError(w, err)
		return
	}

	if value, err = h.extractInt64(req); err != nil {
		handleAppError(w, err)
		return
	}

	metrics := &domain.Metrics{
		ID:    name,
		MType: domain.CounterType,
		Delta: domain.DeltaPtr(value),
	}

	if _, err := h.metricApp.Update(req.Context(), metrics); err != nil {
		handleAppError(w, err)
		return
	}
}

func (h *metricOperationAdapter) GetCounter(w http.ResponseWriter, req *http.Request) {
	_, _ = io.ReadAll(req.Body)
	defer req.Body.Close()

	w.Header().Set("Content-Type", TextPlain)

	var name string
	var err error

	if name, err = h.extractName(req); err != nil {
		handleAppError(w, err)
		return
	}

	value, err := h.metricApp.Get(req.Context(), domain.CounterType, name)
	if err != nil {
		handleAppError(w, err)
		return
	}

	if value == nil {
		err := fmt.Errorf("%w: unknown metric '%v'", domain.ErrNotFound, name)
		handleAppError(w, err)
		return
	}

	if _, err := w.Write([]byte(fmt.Sprintf("%v", *value.Delta))); err != nil {
		handleAppError(w, err)
		return
	}
}

func (h *metricOperationAdapter) GetGauge(w http.ResponseWriter, req *http.Request) {
	_, _ = io.ReadAll(req.Body)
	defer req.Body.Close()

	w.Header().Set("Content-Type", TextPlain)

	var name string
	var err error

	if name, err = h.extractName(req); err != nil {
		handleAppError(w, err)
		return
	}

	value, err := h.metricApp.Get(req.Context(), domain.GaugeType, name)
	if err != nil {
		handleAppError(w, err)
		return
	}

	if value == nil {
		err := fmt.Errorf("%w: unknown metric '%v'", domain.ErrNotFound, name)
		handleAppError(w, err)
		return
	}

	if _, err := w.Write([]byte(fmt.Sprintf("%v", *value.Value))); err != nil {
		handleAppError(w, err)
		return
	}
}

func (h *metricOperationAdapter) AllMetrics(w http.ResponseWriter, req *http.Request) {
	_, _ = io.ReadAll(req.Body)
	defer req.Body.Close()

	metricses, err := h.metricApp.GetAllMetrics(req.Context())
	if err != nil {
		handleAppError(w, err)
		return
	}

	w.Header().Set("Content-Type", TextHTML)

	if err := allMetricsViewTmplate.Execute(w, metricses); err != nil {
		fullErr := fmt.Errorf("%w: generate result error - %v", domain.ErrServerInternal, err.Error())
		handleAppError(w, fullErr)
		return
	}
}

var allMetricsViewTmplate, _ = template.New("allMetrics").Parse(`<!DOCTYPE html>
<html lang="en">
<body>
<table>
    <tr>
        <th>Type</th>
        <th>Name</th>
        <th>Value</th>
    </tr>
    {{ range .}}<tr>
        <td>{{ .MType }}</td>
        <td>{{ .ID }}</td>
        {{if .Delta}}<td>{{ .Delta }}</td>{{else}}<td>{{ .Value }}</td>{{end}}
    </tr>{{ end}}
</table>
</body>
</html>
`)

func (h *metricOperationAdapter) sendMetrics(w http.ResponseWriter, metrics *domain.Metrics) error {
	w.Header().Set("Content-Type", ApplicationJSON)
	if resp, err := json.Marshal(metrics); err != nil {
		return err
	} else {
		w.Write(resp)
		return nil
	}
}

func (h *metricOperationAdapter) checkContentType(expectedType string, req *http.Request) error {
	contentType := req.Header.Get("Content-Type")
	if contentType != "" && !strings.HasPrefix(contentType, expectedType) {
		return fmt.Errorf("%w: only '%v' supported", ErrMediaType, expectedType)
	}
	return nil
}

func (h *metricOperationAdapter) extractName(req *http.Request) (string, error) {
	name := chi.URLParam(req, "name")
	if name == "" {
		err := fmt.Errorf("%w: name is not set", domain.ErrDataFormat)
		return "", err
	}
	return name, nil
}

func (h *metricOperationAdapter) extractFloat64(req *http.Request) (float64, error) {

	valueStr := chi.URLParam(req, "value")
	value, err := domain.ExtractFloat64(valueStr)
	if err != nil {
		fullErr := fmt.Errorf("%w: extract float64 error - %v", domain.ErrDataFormat, err.Error())
		return 0, fullErr
	}
	return value, nil
}

func (h *metricOperationAdapter) extractInt64(req *http.Request) (int64, error) {
	valueStr := chi.URLParam(req, "value")
	value, err := domain.ExtractInt64(valueStr)
	if err != nil {
		fullErr := fmt.Errorf("%w: extract int64 error - %v", domain.ErrDataFormat, err.Error())
		return 0, fullErr
	}
	return value, nil
}
