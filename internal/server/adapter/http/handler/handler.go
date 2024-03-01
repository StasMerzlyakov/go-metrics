package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type MetricUseCase interface {
	GetAllMetrics() ([]domain.Metrics, error)
	GetCounter(name string) (*domain.Metrics, error)
	GetGauge(name string) (*domain.Metrics, error)
	AddCounter(m *domain.Metrics) error
	SetGauge(m *domain.Metrics) error
}

func NewHTTP(metricUseCase MetricUseCase, log *zap.SugaredLogger, middlewares ...func(http.Handler) http.Handler) http.Handler {
	return createHTTPHandler(&httpAdapter{
		metricController: metricUseCase,
		logger:           log,
	}, middlewares...)
}

func createHTTPHandler(httpAdapter *httpAdapter, middlewares ...func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()

	r.Use(middlewares...)

	r.Get("/", httpAdapter.AllMetrics)

	r.Route("/update", func(r chi.Router) {
		r.Post("/", httpAdapter.PostMetrics)
		r.Post("/gauge/{name}/{value}", httpAdapter.PostGauge)
		r.Post("/gauge/{name}", StatusNotFound)
		r.Post("/counter/{name}/{value}", httpAdapter.PostCounter)
		r.Post("/counter/{name}", StatusNotFound)
		r.Post("/{type}/{name}/{value}", StatusNotImplemented)
	})

	r.Route("/value", func(r chi.Router) {
		r.Post("/", httpAdapter.ValueMetrics)
		r.Get("/gauge/{name}", httpAdapter.GetGauge)
		r.Get("/counter/{name}", httpAdapter.GetCounter)
	})
	return r
}

// TODO много повторяющегося кода
type httpAdapter struct {
	metricController MetricUseCase
	logger           *zap.SugaredLogger
}

func (h *httpAdapter) PostGauge(w http.ResponseWriter, req *http.Request) {
	_, _ = io.ReadAll(req.Body)
	// Проверка content-type
	contentType := req.Header.Get("Content-Type")
	if contentType != "" && !strings.HasPrefix(contentType, TextPlain) {
		h.logger.Infow("PostGauge", "status", "error", "msg", fmt.Sprintf("unexpected content-type: %v", contentType))
		http.Error(w, "only 'text/plain' supported", http.StatusUnsupportedMediaType)
		return
	}

	// Извлекаем имя
	name := chi.URLParam(req, "name")
	if name == "" {
		err := errors.New("name is not set")
		h.logger.Infow("PostGauge", "status", "error", "msg", "err", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Проверка допустимости значения параметра
	valueStr := chi.URLParam(req, "value")
	value, err := ExtractFloat64(valueStr)
	if err != nil {
		h.logger.Infow("PostGauge", "status", "error", "msg", "err", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	metrics := &domain.Metrics{
		ID:    name,
		MType: domain.GaugeType,
		Value: domain.ValuePtr(value),
	}

	if err := h.metricController.SetGauge(metrics); err != nil {
		h.handlerStorageError("PostGauge", err, w)
		return
	}
	h.logger.Infow("PostGauge", "name", name, "status", "ok")
	w.WriteHeader(http.StatusOK)
}

func (h *httpAdapter) GetGauge(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", TextPlain)
	// Извлекаем имя
	name := chi.URLParam(req, "name")
	if name == "" {
		err := errors.New("name is not set")
		h.logger.Infow("GetGauge", "status", "error", "msg", "err", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	m, err := h.metricController.GetGauge(name)
	if err != nil {
		h.handlerStorageError("GetGauge", err, w)
		return
	}

	if m == nil {
		h.logger.Infow("GetGauge", "status", "error", "msg", fmt.Sprintf("can't find gague by name '%v'", name))
		w.WriteHeader(http.StatusNotFound)
	} else {
		h.logger.Infow("GetGauge", "name", name, "status", "ok")
		w.Write([]byte(fmt.Sprintf("%v", *m.Value)))
	}
}

func (h *httpAdapter) PostCounter(w http.ResponseWriter, req *http.Request) {
	_, _ = io.ReadAll(req.Body)

	// Проверка content-type
	contentType := req.Header.Get("Content-Type")
	if contentType != "" && !strings.HasPrefix(contentType, TextPlain) {
		h.logger.Infow("PostCounter", "status", "error", "msg", fmt.Sprintf("unexpected content-type: %v", contentType))
		http.Error(w, "only 'text/plain' supported", http.StatusUnsupportedMediaType)
		return
	}

	// Извлекаем имя
	name := chi.URLParam(req, "name")
	if name == "" {
		err := errors.New("name is not set")
		h.logger.Infow("PostGauge", "status", "error", "msg", "err", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Проверка допустимости значения
	valueStr := chi.URLParam(req, "value")

	value, err := ExtractInt64(valueStr)
	if err != nil {
		h.logger.Infow("PostCounter", "status", "error", "msg", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	metrics := &domain.Metrics{
		ID:    name,
		MType: domain.CounterType,
		Delta: domain.DeltaPtr(value),
	}

	if err := h.metricController.AddCounter(metrics); err != nil {
		h.handlerStorageError("PostCounter", err, w)
		return
	}

	h.logger.Infow("PostCounter", "name", name, "status", "ok")
}

func (h *httpAdapter) GetCounter(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", TextPlain)
	// Извлекаем имя
	name := chi.URLParam(req, "name")
	if name == "" {
		err := errors.New("name is not set")
		h.logger.Infow("GetCounter", "status", "error", "msg", "err", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	m, err := h.metricController.GetCounter(name)
	if err != nil {
		h.handlerStorageError("GetCounter", err, w)
		return
	}

	if m == nil {
		h.logger.Infow("GetCounter", "status", "error", "msg", fmt.Sprintf("can't find counter by name '%v'", name))
		w.WriteHeader(http.StatusNotFound)
	} else {
		h.logger.Infow("GetCounter", "name", name, "status", "ok")
		w.Write([]byte(fmt.Sprintf("%v", *m.Delta)))
	}
}

func (h *httpAdapter) AllMetrics(w http.ResponseWriter, request *http.Request) {
	metricses, err := h.metricController.GetAllMetrics()
	if err != nil {
		h.handlerStorageError("AllMetrics", err, w)
		return
	}

	w.Header().Set("Content-Type", TextHTML)

	allMetricsViewTmpl.Execute(w, metricses)
	h.logger.Infow("AllMetrics", "status", "ok")
}

var allMetricsViewTmpl, _ = template.New("allMetrics").Parse(`<!DOCTYPE html>
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

func (h *httpAdapter) PostMetrics(w http.ResponseWriter, req *http.Request) {

	if metrics, ok := h.checkRequestBody("PostMetrics", w, req); ok {
		var err error
		if metrics.MType == "gauge" {
			err = h.metricController.SetGauge(metrics)
		}
		if metrics.MType == "counter" {
			err = h.metricController.AddCounter(metrics)
		}

		if err != nil {
			h.handlerStorageError("PostMetrics", err, w)
			return
		}
		if err := h.sendMetrics("PostMetrics", w, metrics); err == nil {
			h.logger.Infow("PostMetrics", "name", metrics.ID, "status", "ok")
		}
	}
}

func (h *httpAdapter) ValueMetrics(w http.ResponseWriter, req *http.Request) {
	if metrics, ok := h.checkRequestBody("ValueMetrics", w, req); ok {
		var err error
		var metricsFound *domain.Metrics

		if metrics.MType == "gauge" {
			metricsFound, err = h.metricController.GetGauge(metrics.ID)
		}

		if metrics.MType == "counter" {
			metricsFound, err = h.metricController.GetCounter(metrics.ID)
		}

		if err != nil {
			h.handlerStorageError("ValueMetrics", err, w)
			return
		}

		if metricsFound == nil {
			h.logger.Infow("ValueMetrics", "status", "error", "msg",
				fmt.Sprintf("can't find metrics by ID '%v' and MType '%v'", metrics.ID, metrics.MType))
			w.WriteHeader(http.StatusNotFound)
		} else {
			if err := h.sendMetrics("ValueMetrics", w, metricsFound); err == nil {
				h.logger.Infow("ValueMetrics", "name", metrics.ID, "status", "ok")
			}
		}
	}
}

func (h *httpAdapter) handlerStorageError(action string, err error, w http.ResponseWriter) {
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

func (h *httpAdapter) checkRequestBody(action string, w http.ResponseWriter, req *http.Request) (*domain.Metrics, bool) {
	// Проверка content-type
	contentType := req.Header.Get("Content-Type")
	if contentType != "" && !strings.HasPrefix(contentType, ApplicationJSON) {
		h.logger.Infow(action, "status", "error", "msg", fmt.Sprintf("unexpected content-type: %v", contentType))
		http.Error(w, "only 'text/plain' supported", http.StatusUnsupportedMediaType)
		return nil, false
	}

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

func (h *httpAdapter) sendMetrics(action string, w http.ResponseWriter, metrics *domain.Metrics) error {
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

func ExtractFloat64(valueStr string) (float64, error) {
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return -1, err
	}
	return value, nil
}

func ExtractInt64(valueStr string) (int64, error) {
	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		return -1, err
	}
	return value, nil
}

// TODO дублируется с internal/server/server.go
const (
	ApplicationJSON = "application/json"
	TextPlain       = "text/plain"
	TextHTML        = "text/html"
)

func BadRequestHandler(w http.ResponseWriter, req *http.Request) {
	http.Error(w, "BadRequest", http.StatusBadRequest)
}

func StatusMethodNotAllowedHandler(w http.ResponseWriter, req *http.Request) {
	http.Error(w, "StatusMethodNotAllowed", http.StatusMethodNotAllowed)
}
func StatusNotImplemented(w http.ResponseWriter, req *http.Request) {
	http.Error(w, "StatusMethodNotAllowed", http.StatusNotImplemented)
}

func StatusNotFound(w http.ResponseWriter, req *http.Request) {
	http.Error(w, "StatusNotFound", http.StatusNotFound)
}

func TodoResponse(res http.ResponseWriter, message string) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusNotImplemented)
	fmt.Fprintf(res, `
      {
        "response": {
          "text": "%v"
        },
        "version": "1.0"
      }
    `, message)
}
