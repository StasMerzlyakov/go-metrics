package server

import (
	"fmt"
	"github.com/StasMerzlyakov/go-metrics/internal"
	"github.com/StasMerzlyakov/go-metrics/internal/storage"
	"github.com/go-chi/chi/v5"
	"html/template"
	"net/http"
	"strings"
)

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

func CheckMethodPostMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(w, "only post methods", http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, req)
	}
}

func CheckContentTypeMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		contentType := req.Header.Get("Content-Type")
		if contentType != "" && !strings.HasPrefix(contentType, "text/plain") {
			http.Error(w, "only 'text/plain' supported", http.StatusUnsupportedMediaType)
			return
		}
		next.ServeHTTP(w, req)
	}
}

func CheckDigitalMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		valueStr := chi.URLParam(req, "value")
		if !CheckDecimal(valueStr) {
			http.Error(w, "wrong decimal value", http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, req)
	}
}

func CheckIntegerMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		valueStr := chi.URLParam(req, "value")
		if !CheckInteger(valueStr) {
			http.Error(w, "wrong integer value", http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, req)
	}
}

func CheckMetricNameMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		name := chi.URLParam(req, "name")
		if !CheckName(name) {
			http.Error(w, "wrong name value", http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, req)
	}
}

func GaugePostHandlerCreator(storage storage.MetricsStorage[float64]) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		name := chi.URLParam(req, "name")
		valueStr := chi.URLParam(req, "value")
		value, err := ExtractFloat64(valueStr)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		storage.Set(name, value)
		res.WriteHeader(http.StatusOK)
	}
}

func GaugeGetHandlerCreator(gaugeStorage storage.MetricsStorage[float64]) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		name := chi.URLParam(req, "name")
		if v, ok := gaugeStorage.Get(name); !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		} else {
			w.Write([]byte(fmt.Sprintf("%v", v)))
		}
		w.WriteHeader(http.StatusOK)
	}
}

func CounterPostHandlerCreator(storage storage.MetricsStorage[int64]) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		name := chi.URLParam(req, "name")
		valueStr := chi.URLParam(req, "value")
		value, err := ExtractInt64(valueStr)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		storage.Add(name, value)
		res.WriteHeader(http.StatusOK)
	}
}

func CounterGetHandlerCreator(counterStorage storage.MetricsStorage[int64]) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		name := chi.URLParam(req, "name")
		if v, ok := counterStorage.Get(name); !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		} else {
			w.Write([]byte(fmt.Sprintf("%v", v)))
		}
		w.WriteHeader(http.StatusOK)
	}
}

type MetricsData struct {
	Type  string
	Name  string
	Value string
}

type MetricModel struct {
	Items []MetricsData
}

// AllMetricsViewHandlerCreator
// по мотивам https://stackoverflow.com/questions/56923511/go-html-template-table
func AllMetricsViewHandlerCreator(
	counterStorage storage.MetricsStorage[int64],
	gaugeStorage storage.MetricsStorage[float64]) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		metrics := GetAllMetrics(counterStorage, gaugeStorage)
		allMetricsViewTmpl.Execute(w, metrics)
	}
}

var DefaultHandler = func(w http.ResponseWriter, req *http.Request) {
	http.Error(w, "", http.StatusNotImplemented)
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
