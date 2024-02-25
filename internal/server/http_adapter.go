package server

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type MetricController interface {
	GetAllMetrics() MetricModel
	GetCounter(name string) (int64, bool)
	GetGaguge(name string) (float64, bool)
	AddCounter(name string, value int64)
	SetGauge(name string, value float64)
}

func NewHttpAdapterHandler(metricController MetricController, log *zap.SugaredLogger) *httpAdapter {
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
	if contentType != "" && !strings.HasPrefix(contentType, "text/plain") {
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
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
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
	if contentType != "" && !strings.HasPrefix(contentType, "text/plain") {
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
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
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
