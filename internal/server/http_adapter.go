package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

const (
	applicationJSON = "application/json"
	textPlain       = "text/plain"
	textHTML        = "text/html"
)

type MetricController interface {
	GetAllMetrics() []Metrics
	GetCounter(name string) (int64, bool)
	GetGauge(name string) (float64, bool)
	AddCounter(name string, value int64)
	SetGauge(name string, value float64)
}

func NewHTTPAdapterHandler(metricController MetricController, log *zap.SugaredLogger) *httpAdapter {
	return &httpAdapter{
		metricController: metricController,
		logger:           log,
	}
}

type httpAdapter struct {
	metricController MetricController
	logger           *zap.SugaredLogger
}

func (h *httpAdapter) PostGauge(w http.ResponseWriter, req *http.Request) {
	_, _ = io.ReadAll(req.Body)
	// Проверка content-type
	contentType := req.Header.Get("Content-Type")
	if contentType != "" && !strings.HasPrefix(contentType, textPlain) {
		h.logger.Infow("PostGauge", "status", "error", "msg", fmt.Sprintf("unexpected content-type: %v", contentType))
		http.Error(w, "only 'text/plain' supported", http.StatusUnsupportedMediaType)
		return
	}

	// Проверка допустимости имени параметра
	name := chi.URLParam(req, "name")
	if !CheckName(name) {
		h.logger.Infow("PostGauge", "status", "error", "msg", fmt.Sprintf("wrong value name: %v", name))
		http.Error(w, "wrong name value", http.StatusBadRequest)
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
	h.metricController.SetGauge(name, value)
	h.logger.Infow("PostGauge", "name", name, "status", "ok")
	w.WriteHeader(http.StatusOK)
}

func (h *httpAdapter) GetGauge(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", textPlain)
	name := chi.URLParam(req, "name")
	if v, ok := h.metricController.GetGauge(name); !ok {
		h.logger.Infow("GetGauge", "status", "error", "msg", fmt.Sprintf("can't find gauge by name '%v'", name))
		w.WriteHeader(http.StatusNotFound)
		return
	} else {
		h.logger.Infow("GetGauge", "name", name, "status", "ok")
		w.Write([]byte(fmt.Sprintf("%v", v)))
	}
}

func (h *httpAdapter) PostCounter(w http.ResponseWriter, req *http.Request) {
	_, _ = io.ReadAll(req.Body)

	// Проверка content-type
	contentType := req.Header.Get("Content-Type")
	if contentType != "" && !strings.HasPrefix(contentType, textPlain) {
		h.logger.Infow("PostCounter", "status", "error", "msg", fmt.Sprintf("unexpected content-type: %v", contentType))
		http.Error(w, "only 'text/plain' supported", http.StatusUnsupportedMediaType)
		return
	}

	// Проверка допустимости имени параметра
	name := chi.URLParam(req, "name")
	if !CheckName(name) {
		h.logger.Infow("PostCounter", "status", "error", "msg", fmt.Sprintf("wrong value name: %v", name))
		http.Error(w, "wrong name value", http.StatusBadRequest)
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
	h.metricController.AddCounter(name, value)
	h.logger.Infow("PostCounter", "name", name, "status", "ok")
}

func (h *httpAdapter) GetCounter(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", textPlain)
	name := chi.URLParam(req, "name")
	if v, ok := h.metricController.GetCounter(name); !ok {
		h.logger.Infow("GetCounter", "status", "error", "msg", fmt.Sprintf("can't find counter by name '%v'", name))
		w.WriteHeader(http.StatusNotFound)
	} else {
		h.logger.Infow("GetCounter", "name", name, "status", "ok")
		w.Write([]byte(fmt.Sprintf("%v", v)))
	}
}

func (h *httpAdapter) AllMetrics(w http.ResponseWriter, request *http.Request) {
	metrics := h.metricController.GetAllMetrics()
	w.Header().Set("Content-Type", textHTML)

	allMetricsViewTmpl.Execute(w, metrics)
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
    {{ range .}}
        <tr>
            <td>{{ .MType }}</td>
            <td>{{ .ID }}</td>
			{{if .Delta}}
			<td>{{ .Delta }}</td>
			{{else}}
			<td>{{ .Value }}</td>
			{{end}}
            
        </tr>
    {{ end}}
</table>
</body>
</html>
`)

func (h *httpAdapter) checkJSONInput(msg string, w http.ResponseWriter, req *http.Request) (*Metrics, bool) {
	// Проверка content-type
	contentType := req.Header.Get("Content-Type")
	if contentType != "" && !strings.HasPrefix(contentType, applicationJSON) {
		h.logger.Infow(msg, "status", "error", "msg", fmt.Sprintf("unexpected content-type: %v", contentType))
		http.Error(w, "only 'text/plain' supported", http.StatusUnsupportedMediaType)
		return nil, false
	}

	// Декодируем входные данные
	var metrics Metrics
	if err := json.NewDecoder(req.Body).Decode(&metrics); err != nil {
		h.logger.Infow(msg, "status", "error", "msg", fmt.Sprintf("json decode error: %v", err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, false
	}

	// Проверка типа
	if metrics.MType != "counter" && metrics.MType != "gauge" {
		errMess := fmt.Sprintf("unexpected MType %v, expected 'counter' and 'gauge'", metrics.MType)
		h.logger.Infow(msg, "status", "error", "msg", errMess)
		http.Error(w, errMess, http.StatusBadRequest)
		return nil, false
	}

	return &metrics, true
}

func (h *httpAdapter) PostMetric(w http.ResponseWriter, req *http.Request) {

	if metrics, ok := h.checkJSONInput("PostMetric", w, req); ok {
		if metrics.MType == "gauge" {
			if metrics.Value == nil {
				h.logger.Infow("PostMetric", "status", "error", "msg", "gague value is nil")
				http.Error(w, "gague value is nil", http.StatusBadRequest)
				return
			}
			h.metricController.SetGauge(metrics.ID, *metrics.Value)
			value, _ := h.metricController.GetGauge(metrics.ID)
			metrics.Value = &value
		}

		if metrics.MType == "counter" {
			if metrics.Delta == nil {
				h.logger.Infoln("PostMetric", "status", "error", "msg", "counter delta is nil")
				http.Error(w, "counter delta is nil", http.StatusBadRequest)
				return
			}
			h.metricController.AddCounter(metrics.ID, *metrics.Delta)
			value, _ := h.metricController.GetCounter(metrics.ID)
			metrics.Delta = &value
		}

		h.sendMetrics(w, req, metrics)
		h.logger.Infow("PostMetric", "name", metrics.ID, "status", "ok")
	}
}

func (h *httpAdapter) sendMetrics(w http.ResponseWriter, req *http.Request, metrics *Metrics) {
	w.Header().Set("Content-Type", applicationJSON)
	resp, err := json.Marshal(metrics)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(resp)
}

func (h *httpAdapter) ValueMetric(w http.ResponseWriter, req *http.Request) {
	if metrics, ok := h.checkJSONInput("ValueMetric", w, req); ok {
		if metrics.MType == "gauge" {
			value, ok := h.metricController.GetGauge(metrics.ID)
			if !ok {
				errMess := fmt.Sprintf("metric by name %v and type %v not found", metrics.ID, metrics.MType)
				h.logger.Infow("ValueMetric", "status", "error", "msg", errMess)
				http.Error(w, errMess, http.StatusNotFound)
				return
			}
			metrics.Value = &value
		}

		if metrics.MType == "counter" {
			value, ok := h.metricController.GetCounter(metrics.ID)
			if !ok {
				errMess := fmt.Sprintf("metric by name %v and type %v not found", metrics.ID, metrics.MType)
				h.logger.Infow("ValueMetric", "status", "error", "msg", errMess)
				http.Error(w, errMess, http.StatusNotFound)
				return
			}
			metrics.Delta = &value
		}

		h.sendMetrics(w, req, metrics)
		h.logger.Infow("ValueMetric", "msg", fmt.Sprintf("metric by name %v found", metrics.ID), "status", "ok")
	}
}
