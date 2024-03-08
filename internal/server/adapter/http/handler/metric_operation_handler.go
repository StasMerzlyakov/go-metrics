package handler

import (
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
)

func AddMetricOperations(r *chi.Mux, metricApp MetricApp, log *zap.SugaredLogger) {

	adapter := &metricOperationAdapter{
		metricApp: metricApp,
		logger:    log,
	}

	r.Get("/", adapter.AllMetrics)

	r.Route("/update", func(r chi.Router) {
		r.Post("/", adapter.PostMetrics)
		r.Post("/gauge/{name}/{value}", adapter.PostGauge)
		r.Post("/gauge/{name}", StatusNotFound)
		r.Post("/counter/{name}/{value}", adapter.PostCounter)
		r.Post("/counter/{name}", StatusNotFound)
		r.Post("/{type}/{name}/{value}", StatusNotImplemented)
	})

	r.Route("/value", func(r chi.Router) {
		r.Post("/", adapter.ValueMetrics)
		r.Get("/gauge/{name}", adapter.GetGauge)
		r.Get("/counter/{name}", adapter.GetCounter)
	})
}

type metricOperationAdapter struct {
	metricApp MetricApp
	logger    *zap.SugaredLogger
}

func (h *metricOperationAdapter) checkRequestBody(w http.ResponseWriter, req *http.Request) (*domain.Metrics, bool) {

	action := h.getAction()

	// Декодируем входные данные
	var metrics domain.Metrics
	if err := json.NewDecoder(req.Body).Decode(&metrics); err != nil {
		h.logger.Infow(action, "status", "error", "msg", fmt.Sprintf("json decode error: %v", err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, false
	}

	// Проверка типа
	if metrics.MType != "counter" && metrics.MType != "gauge" {
		errMess := fmt.Sprintf("unexpected MType %v, expected 'counter' and 'gauge'", metrics.MType)
		h.logger.Infow(action, "status", "error", "msg", errMess)
		http.Error(w, errMess, http.StatusBadRequest)
		return nil, false
	}

	return &metrics, true
}

func (h *metricOperationAdapter) PostMetrics(w http.ResponseWriter, req *http.Request) {
	action := "PostMetrics"

	if !h.isContentTypeExpected(ApplicationJSON, w, req) {
		return
	}

	if metrics, ok := h.checkRequestBody(w, req); ok {
		var err error
		if metrics.MType == "gauge" {
			err = h.metricApp.SetGauge(req.Context(), metrics)
		}
		if metrics.MType == "counter" {
			err = h.metricApp.AddCounter(req.Context(), metrics)
		}

		if err != nil {
			h.handlerAppError(err, w)
			return
		}
		if err := h.sendMetrics(w, metrics); err == nil {
			h.logger.Infow(action, "name", metrics.ID, "status", "ok")
		}
	}
}

func (h *metricOperationAdapter) ValueMetrics(w http.ResponseWriter, req *http.Request) {

	action := "ValueMetrics"

	if !h.isContentTypeExpected(ApplicationJSON, w, req) {
		return
	}

	if metrics, ok := h.checkRequestBody(w, req); ok {
		var err error
		var metricsFound *domain.Metrics

		if metrics.MType == "gauge" {
			metricsFound, err = h.metricApp.GetGauge(req.Context(), metrics.ID)
		}

		if metrics.MType == "counter" {
			metricsFound, err = h.metricApp.GetCounter(req.Context(), metrics.ID)
		}

		if err != nil {
			h.handlerAppError(err, w)
			return
		}

		if metricsFound == nil {
			h.logger.Infow(action, "status", "error", "msg",
				fmt.Sprintf("can't find metrics by ID '%v' and MType '%v'", metrics.ID, metrics.MType))
			w.WriteHeader(http.StatusNotFound)
		} else {
			if err := h.sendMetrics(w, metricsFound); err == nil {
				h.logger.Infow(action, "name", metrics.ID, "status", "ok")
			}
		}
	}
}

func (h *metricOperationAdapter) PostGauge(w http.ResponseWriter, req *http.Request) {

	opName := "PostGauge"
	_, _ = io.ReadAll(req.Body)

	if !h.isContentTypeExpected(TextPlain, w, req) {
		return
	}

	var name string
	var value float64
	var ok bool

	if name, ok = h.extractName(w, req); !ok {
		return
	}

	if value, ok = h.extractFloat64(w, req); !ok {
		return
	}

	metrics := &domain.Metrics{
		ID:    name,
		MType: domain.GaugeType,
		Value: domain.ValuePtr(value),
	}

	if err := h.metricApp.SetGauge(req.Context(), metrics); err != nil {
		h.handlerAppError(err, w)
		return
	}
	h.logger.Infow(opName, "name", name, "status", "ok")
}

func (h *metricOperationAdapter) PostCounter(w http.ResponseWriter, req *http.Request) {

	opName := "PostCounter"
	_, _ = io.ReadAll(req.Body)

	if !h.isContentTypeExpected(TextPlain, w, req) {
		return
	}

	var name string
	var value int64
	var ok bool

	if name, ok = h.extractName(w, req); !ok {
		return
	}

	if value, ok = h.extractInt64(w, req); !ok {
		return
	}

	metrics := &domain.Metrics{
		ID:    name,
		MType: domain.CounterType,
		Delta: domain.DeltaPtr(value),
	}

	if err := h.metricApp.AddCounter(req.Context(), metrics); err != nil {
		h.handlerAppError(err, w)
		return
	}

	h.logger.Infow(opName, "name", name, "status", "ok")
}

func (h *metricOperationAdapter) GetCounter(w http.ResponseWriter, req *http.Request) {
	action := "GetCounter"
	w.Header().Set("Content-Type", TextPlain)

	var name string
	var ok bool
	if name, ok = h.extractName(w, req); !ok {
		return
	}

	m, err := h.metricApp.GetCounter(req.Context(), name)
	if err != nil {
		h.handlerAppError(err, w)
		return
	}

	if m == nil {
		h.logger.Infow(action, "status", "error", "msg", fmt.Sprintf("can't find counter by name '%v'", name))
		w.WriteHeader(http.StatusNotFound)
	} else {
		h.logger.Infow(action, "name", name, "status", "ok")
		w.Write([]byte(fmt.Sprintf("%v", *m.Delta)))
	}
}

func (h *metricOperationAdapter) GetGauge(w http.ResponseWriter, req *http.Request) {
	action := "GetGauge"

	w.Header().Set("Content-Type", TextPlain)
	var name string
	var ok bool

	if name, ok = h.extractName(w, req); !ok {
		return
	}

	m, err := h.metricApp.GetGauge(req.Context(), name)
	if err != nil {
		h.handlerAppError(err, w)
		return
	}

	if m == nil {
		h.logger.Infow(action, "status", "error", "msg", fmt.Sprintf("can't find gague by name '%v'", name))
		w.WriteHeader(http.StatusNotFound)
	} else {
		h.logger.Infow(action, "name", name, "status", "ok")
		w.Write([]byte(fmt.Sprintf("%v", *m.Value)))
	}
}

func (h *metricOperationAdapter) AllMetrics(w http.ResponseWriter, request *http.Request) {
	metricses, err := h.metricApp.GetAllMetrics(request.Context())
	if err != nil {
		h.handlerAppError(err, w)
		return
	}

	w.Header().Set("Content-Type", TextHTML)

	allMetricsViewTmplate.Execute(w, metricses)
	h.logger.Infow("AllMetrics", "status", "ok")
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
	action := h.getAction()
	w.Header().Set("Content-Type", ApplicationJSON)
	resp, err := json.Marshal(metrics)
	if err != nil {
		h.logger.Infow(action, "status", "error", "msg", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	w.Write(resp)
	return nil
}

func (h *metricOperationAdapter) isContentTypeExpected(expectedType string, w http.ResponseWriter, req *http.Request) bool {
	action := h.getAction()

	contentType := req.Header.Get("Content-Type")
	if contentType != "" && !strings.HasPrefix(contentType, expectedType) {
		h.logger.Infow(action, "status", "error", "msg", fmt.Sprintf("unexpected content-type: %v", contentType))
		http.Error(w, fmt.Sprintf("only '%v' supported", expectedType), http.StatusUnsupportedMediaType)
		return false
	}
	return true
}

func (h *metricOperationAdapter) extractName(w http.ResponseWriter, req *http.Request) (string, bool) {
	action := h.getAction()

	name := chi.URLParam(req, "name")
	if name == "" {
		err := errors.New("name is not set")
		h.logger.Infow(action, "status", "error", "msg", "err", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return "", false
	}
	return name, true
}

func (h *metricOperationAdapter) extractFloat64(w http.ResponseWriter, req *http.Request) (float64, bool) {
	action := h.getAction()

	valueStr := chi.URLParam(req, "value")
	value, err := domain.ExtractFloat64(valueStr)
	if err != nil {
		h.logger.Infow(action, "status", "error", "msg", "err", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return 0, false
	}
	return value, true
}

func (h *metricOperationAdapter) extractInt64(w http.ResponseWriter, req *http.Request) (int64, bool) {
	action := h.getAction()

	valueStr := chi.URLParam(req, "value")
	value, err := domain.ExtractInt64(valueStr)
	if err != nil {
		h.logger.Infow(action, "status", "error", "msg", "err", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return 0, false
	}
	return value, true
}

func (h *metricOperationAdapter) getAction() string {
	// используется для получения имени вызывющей функции
	// по мотивам https://stackoverflow.com/questions/25927660/how-to-get-the-current-function-name
	pc, _, _, _ := runtime.Caller(2)
	op := runtime.FuncForPC(pc).Name()
	return op
}

func (h *metricOperationAdapter) handlerAppError(err error, w http.ResponseWriter) {
	pc, _, _, _ := runtime.Caller(1)
	action := runtime.FuncForPC(pc).Name()

	if errors.Is(err, domain.ErrDataFormat) {
		h.logger.Infow(action, "status", "error", "msg", "err", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else {
		h.logger.Infow(action, "status", "error", "msg", "err", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
