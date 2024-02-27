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
)

type MetricController interface {
	GetAllMetrics() MetricModel
	GetCounter(name string) (int64, bool)
	GetGaguge(name string) (float64, bool)
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
		h.logger.Infoln("err", fmt.Sprintf("unexpected content-type: %v", contentType))
		http.Error(w, "only 'text/plain' supported", http.StatusUnsupportedMediaType)
		return
	}

	// Проверка допустимости имени параметра
	name := chi.URLParam(req, "name")
	if !CheckName(name) {
		h.logger.Infoln("err", fmt.Sprintf("wrong value name: %v", name))
		http.Error(w, "wrong name value", http.StatusBadRequest)
		return
	}

	// Проверка допустимости значения параметра
	valueStr := chi.URLParam(req, "value")
	if !CheckDecimal(valueStr) {
		h.logger.Infoln("err", fmt.Sprintf("wrong decimal value: %v", valueStr))
		http.Error(w, "wrong decimal value", http.StatusBadRequest)
		return
	}

	value, err := ExtractFloat64(valueStr)
	if err != nil {
		h.logger.Infoln("err", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.metricController.SetGauge(name, value)
	w.WriteHeader(http.StatusOK)
}

func (h *httpAdapter) GetGauge(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", textPlain)
	name := chi.URLParam(req, "name")
	if v, ok := h.metricController.GetGaguge(name); !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	} else {
		w.Write([]byte(fmt.Sprintf("%v", v)))
	}
}

func (h *httpAdapter) PostCounter(w http.ResponseWriter, req *http.Request) {
	_, _ = io.ReadAll(req.Body)

	// Проверка content-type
	contentType := req.Header.Get("Content-Type")
	if contentType != "" && !strings.HasPrefix(contentType, textPlain) {
		h.logger.Infoln("err", fmt.Sprintf("unexpected content-type: %v", contentType))
		http.Error(w, "only 'text/plain' supported", http.StatusUnsupportedMediaType)
		return
	}

	// Проверка допустимости имени параметра
	name := chi.URLParam(req, "name")
	if !CheckName(name) {
		h.logger.Infoln("err", fmt.Sprintf("wrong value name: %v", name))
		http.Error(w, "wrong name value", http.StatusBadRequest)
		return
	}

	// Проверка допустимости значения
	valueStr := chi.URLParam(req, "value")
	if !CheckInteger(valueStr) {
		h.logger.Infoln("err", fmt.Sprintf("wrong integer value: %v", valueStr))
		http.Error(w, "wrong integer value", http.StatusBadRequest)
		return
	}

	value, err := ExtractInt64(valueStr)
	if err != nil {
		h.logger.Infoln("err", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.metricController.AddCounter(name, value)
	w.WriteHeader(http.StatusOK)
}

func (h *httpAdapter) GetCounter(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", textPlain)
	name := chi.URLParam(req, "name")
	if v, ok := h.metricController.GetCounter(name); !ok {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.Write([]byte(fmt.Sprintf("%v", v)))
	}
}

func (h *httpAdapter) AllMetrics(w http.ResponseWriter, request *http.Request) {
	metrics := h.metricController.GetAllMetrics()
	allMetricsViewTmpl.Execute(w, metrics)
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
    {{ range .Items}}
        <tr>
            <td>{{ .Type }}</td>
            <td>{{ .Name }}</td>
            <td>{{ .Value }}</td>
        </tr>
    {{ end}}
</table>
</body>
</html>
`)

func (h *httpAdapter) checkJsonInput(w http.ResponseWriter, req *http.Request) (*Metrics, bool) {
	// Проверка content-type
	contentType := req.Header.Get("Content-Type")
	if contentType != "" && !strings.HasPrefix(contentType, applicationJSON) {
		h.logger.Infoln("err", fmt.Sprintf("unexpected content-type: %v", contentType))
		http.Error(w, "only 'text/plain' supported", http.StatusUnsupportedMediaType)
		return nil, false
	}

	// Декодируем входные данные
	var metrics Metrics
	if err := json.NewDecoder(req.Body).Decode(&metrics); err != nil {
		h.logger.Infoln("err", fmt.Sprintf("json decode error: %v", err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, false
	}

	// Проверка типа
	if metrics.MType != "counter" && metrics.MType != "gauge" {
		h.logger.Infoln("err", "unexpected MType")
		http.Error(w, "unexpected MType", http.StatusBadRequest)
		return nil, false
	}

	return &metrics, true
}

func (h *httpAdapter) PostMetric(w http.ResponseWriter, req *http.Request) {

	if metrics, ok := h.checkJsonInput(w, req); ok {
		if metrics.MType == "gauge" {
			if metrics.Value == nil {
				h.logger.Infoln("err", "gague value is nil")
				http.Error(w, "gague value is nil", http.StatusBadRequest)
				return
			}
			h.metricController.SetGauge(metrics.ID, *metrics.Value)
			value, _ := h.metricController.GetGaguge(metrics.ID)
			metrics.Value = &value
		}

		if metrics.MType == "counter" {
			if metrics.Delta == nil {
				h.logger.Infoln("err", "counter delta is nil")
				http.Error(w, "counter delta is nil", http.StatusBadRequest)
				return
			}
			h.metricController.AddCounter(metrics.ID, *metrics.Delta)
			value, _ := h.metricController.GetCounter(metrics.ID)
			metrics.Delta = &value
		}

		h.sendMetrics(w, req, metrics)
	}
}

func (h *httpAdapter) sendMetrics(w http.ResponseWriter, req *http.Request, metrics *Metrics) {
	/*
		// ??? Если переставить после Encode то вылетит http: superfluous response.WriteHeader call from
		w.Header().Set("Content-Type", applicationJson)
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(metrics)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} */

	w.Header().Set("Content-Type", applicationJSON)
	resp, err := json.Marshal(metrics)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (h *httpAdapter) ValueMetric(w http.ResponseWriter, req *http.Request) {
	if metrics, ok := h.checkJsonInput(w, req); ok {
		if metrics.MType == "gauge" {
			value, ok := h.metricController.GetGaguge(metrics.ID)
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			metrics.Value = &value
		}

		if metrics.MType == "counter" {
			value, ok := h.metricController.GetCounter(metrics.ID)
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			metrics.Delta = &value
		}

		h.sendMetrics(w, req, metrics)
	}
}
