package server

import (
	"fmt"
	"html/template"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type MetricController interface {
	GetAllMetrics() MetricModel
	GetCounter(name string) (int64, bool)
	GetGaguge(name string) (float64, bool)
	AddCounter(name string, value int64)
	SetGauge(name string, value float64)
}

func NewHttpAdapterHandler(metricController MetricController) *httpAdapter {
	return &httpAdapter{
		metricController: metricController,
	}
}

type httpAdapter struct {
	metricController MetricController
}

func (h *httpAdapter) PostGauge(w http.ResponseWriter, req *http.Request) {
	_, _ = io.ReadAll(req.Body)
	name := chi.URLParam(req, "name")
	valueStr := chi.URLParam(req, "value")
	value, err := ExtractFloat64(valueStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.metricController.SetGauge(name, value)
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
	name := chi.URLParam(req, "name")
	valueStr := chi.URLParam(req, "value")
	value, err := ExtractInt64(valueStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.metricController.AddCounter(name, value)
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
