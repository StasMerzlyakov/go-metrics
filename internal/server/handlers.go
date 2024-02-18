package server

import (
	"fmt"
	"html/template"
	"io"
	"net/http"

	"github.com/StasMerzlyakov/go-metrics/internal"
	"github.com/go-chi/chi/v5"
)

type BusinessHandler interface {
	PostGauge(w http.ResponseWriter, req *http.Request)
	GetGauge(w http.ResponseWriter, req *http.Request)
	PostCounter(w http.ResponseWriter, req *http.Request)
	GetCounter(w http.ResponseWriter, req *http.Request)
	AllMetrics(w http.ResponseWriter, request *http.Request)
}

func NewBusinessHandler(metricController MetricController) BusinessHandler {
	return &handler{
		metricController: metricController,
	}
}

type handler struct {
	metricController MetricController
}

func (h *handler) PostGauge(w http.ResponseWriter, req *http.Request) {
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

func (h *handler) GetGauge(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	name := chi.URLParam(req, "name")
	if v, ok := h.metricController.GetGaguge(name); !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	} else {
		w.Write([]byte(fmt.Sprintf("%v", v)))
	}
}

func (h *handler) PostCounter(w http.ResponseWriter, req *http.Request) {
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

func (h *handler) GetCounter(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	name := chi.URLParam(req, "name")
	if v, ok := h.metricController.GetCounter(name); !ok {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.Write([]byte(fmt.Sprintf("%v", v)))
	}
}

func (h *handler) AllMetrics(w http.ResponseWriter, request *http.Request) {
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

func CreateFullPostCounterHandler(counterHandler http.HandlerFunc) http.HandlerFunc {
	return internal.Conveyor(
		counterHandler,
		CheckIntegerMiddleware,
		CheckMetricNameMiddleware,
		CheckContentTypeMiddleware,
		CheckMethodPostMiddleware,
	)
}

func CreateFullPostGaugeHandler(gaugeHandler http.HandlerFunc) http.HandlerFunc {
	return internal.Conveyor(
		gaugeHandler,
		CheckDigitalMiddleware,
		CheckMetricNameMiddleware,
		CheckContentTypeMiddleware,
		CheckMethodPostMiddleware,
	)
}
