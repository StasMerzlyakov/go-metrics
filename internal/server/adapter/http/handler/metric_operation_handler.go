package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"runtime"
	"strings"

	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	_ "github.com/golang/mock/gomock"        // обязательно, требуется в сгенерированных mock-файлах,
	_ "github.com/golang/mock/mockgen/model" // обязательно для корректного запуска mockgen
)

var ErrMediaType = errors.New("UnsupportedMediaTypeError") // ошибку определяю здесь - она специфична для

//go:generate mockgen -destination "../mocks/$GOFILE" -package mocks . MetricApp

type MetricApp interface {
	GetAll(ctx context.Context) ([]domain.Metrics, error)
	Get(ctx context.Context, metricType domain.MetricType, name string) (*domain.Metrics, error)
	UpdateAll(ctx context.Context, mtr []domain.Metrics) error
	Update(ctx context.Context, mtr *domain.Metrics) (*domain.Metrics, error)
}

func AddMetricOperations(r *chi.Mux, metricApp MetricApp, log *zap.SugaredLogger) {

	adapter := &metricOperationAdapter{
		metricApp: metricApp,
		logger:    log,
	}

	r.Get("/", adapter.AllMetrics)

	r.Route("/updates", func(r chi.Router) {
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

type metricOperationAdapter struct {
	metricApp MetricApp
	logger    *zap.SugaredLogger
}

func (h *metricOperationAdapter) PostMetrics(w http.ResponseWriter, req *http.Request) {

	if err := h.checkContentType(ApplicationJSON, req); err != nil {
		h.handlerAppError(w, err)
		return
	}

	var metrics []domain.Metrics
	if err := json.NewDecoder(req.Body).Decode(&metrics); err != nil {
		fullErr := fmt.Errorf("%w: json decode error - %v", domain.ErrDataFormat, err.Error())
		h.handlerAppError(w, fullErr)
		return
	}

	if err := h.metricApp.UpdateAll(req.Context(), metrics); err != nil {
		h.handlerAppError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *metricOperationAdapter) PostMetric(w http.ResponseWriter, req *http.Request) {

	if err := h.checkContentType(ApplicationJSON, req); err != nil {
		h.handlerAppError(w, err)
		return
	}

	var metrics *domain.Metrics

	if err := json.NewDecoder(req.Body).Decode(&metrics); err != nil {
		fullErr := fmt.Errorf("%w: json decode error - %v", domain.ErrDataFormat, err.Error())
		h.handlerAppError(w, fullErr)
		return
	}

	updatedMetric, err := h.metricApp.Update(req.Context(), metrics)
	if err != nil {
		h.handlerAppError(w, err)
		return
	}

	if err := h.sendMetrics(w, updatedMetric); err != nil {
		h.handlerAppError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *metricOperationAdapter) ValueMetric(w http.ResponseWriter, req *http.Request) {

	if err := h.checkContentType(ApplicationJSON, req); err != nil {
		h.handlerAppError(w, err)
		return
	}

	var metrics *domain.Metrics
	if err := json.NewDecoder(req.Body).Decode(&metrics); err != nil {
		fullErr := fmt.Errorf("%w: json decode error - %v", domain.ErrDataFormat, err.Error())
		h.handlerAppError(w, fullErr)
		return
	}

	value, err := h.metricApp.Get(req.Context(), metrics.MType, metrics.ID)
	if err != nil {
		h.handlerAppError(w, err)

	}

	if err := h.sendMetrics(w, value); err != nil {
		h.handlerAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *metricOperationAdapter) PostGauge(w http.ResponseWriter, req *http.Request) {

	if err := h.checkContentType(TextPlain, req); err != nil {
		h.handlerAppError(w, err)
		return
	}

	_, _ = io.ReadAll(req.Body)
	defer req.Body.Close()

	var name string
	var value float64
	var err error

	if name, err = h.extractName(req); err != nil {
		h.handlerAppError(w, err)
		return
	}

	if value, err = h.extractFloat64(req); err != nil {
		h.handlerAppError(w, err)
		return
	}

	metrics := &domain.Metrics{
		ID:    name,
		MType: domain.GaugeType,
		Value: domain.ValuePtr(value),
	}

	if _, err := h.metricApp.Update(req.Context(), metrics); err != nil {
		h.handlerAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *metricOperationAdapter) PostCounter(w http.ResponseWriter, req *http.Request) {

	if err := h.checkContentType(TextPlain, req); err != nil {
		h.handlerAppError(w, err)
		return
	}

	_, _ = io.ReadAll(req.Body)
	defer req.Body.Close()

	var name string
	var value int64
	var err error

	if name, err = h.extractName(req); err != nil {
		h.handlerAppError(w, err)
		return
	}

	if value, err = h.extractInt64(req); err != nil {
		h.handlerAppError(w, err)
		return
	}

	metrics := &domain.Metrics{
		ID:    name,
		MType: domain.CounterType,
		Delta: domain.DeltaPtr(value),
	}

	if _, err := h.metricApp.Update(req.Context(), metrics); err != nil {
		h.handlerAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *metricOperationAdapter) GetCounter(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", TextPlain)

	var name string
	var err error

	if name, err = h.extractName(req); err != nil {
		h.handlerAppError(w, err)
		return
	}

	value, err := h.metricApp.Get(req.Context(), domain.CounterType, name)
	if err != nil {
		h.handlerAppError(w, err)
		return
	}

	if _, err := w.Write([]byte(fmt.Sprintf("%v", *value.Delta))); err != nil {
		h.handlerAppError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *metricOperationAdapter) GetGauge(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", TextPlain)

	var name string
	var err error

	if name, err = h.extractName(req); err != nil {
		h.handlerAppError(w, err)
		return
	}

	value, err := h.metricApp.Get(req.Context(), domain.GaugeType, name)
	if err != nil {
		h.handlerAppError(w, err)
		return
	}

	if _, err := w.Write([]byte(fmt.Sprintf("%v", *value.Value))); err != nil {
		h.handlerAppError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *metricOperationAdapter) AllMetrics(w http.ResponseWriter, request *http.Request) {
	metricses, err := h.metricApp.GetAll(request.Context())
	if err != nil {
		h.handlerAppError(w, err)
		return
	}

	if err := allMetricsViewTmplate.Execute(w, metricses); err != nil {
		fullErr := fmt.Errorf("%w: generate result error - %v", domain.ErrServerInternal, err.Error())
		h.handlerAppError(w, fullErr)
		return
	}

	w.Header().Set("Content-Type", TextHTML)
	w.WriteHeader(http.StatusOK)
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

func (h *metricOperationAdapter) handlerAppError(w http.ResponseWriter, err error) {

	// используется для получения имени вызывющей функции
	// по мотивам https://stackoverflow.com/questions/25927660/how-to-get-the-current-function-name
	pc, _, _, _ := runtime.Caller(1)
	action := runtime.FuncForPC(pc).Name()

	if errors.Is(err, ErrMediaType) {
		h.logger.Infow(action, "status", "error", "msg", err.Error())
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}

	if errors.Is(err, domain.ErrDataFormat) {
		h.logger.Infow(action, "status", "error", "msg", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if errors.Is(err, domain.ErrServerInternal) ||
		errors.Is(err, domain.ErrDBConnection) {
		h.logger.Infow(action, "status", "error", "msg", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if errors.Is(err, domain.ErrNotFound) {
		h.logger.Infow(action, "status", "error", "msg", err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
}
