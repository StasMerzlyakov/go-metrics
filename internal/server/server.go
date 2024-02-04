package server

import (
	"fmt"
	"github.com/StasMerzlyakov/go-metrics/internal"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// CreateServer
// TODO configuration
func CreateServer() error {
	counterStorage := NewInt64Storage()
	gaugeStorage := NewFloat64Storage()
	sux := http.NewServeMux()
	sux.HandleFunc("/", internal.BadRequestHandler)
	sux.Handle("/update/gauge/", http.StripPrefix("/update/gauge", CreateGaugeConveyor(gaugeStorage)))
	sux.Handle("/update/counter/", http.StripPrefix("/update/counter", CreateCounterConveyor(counterStorage)))

	return http.ListenAndServe(`:8080`, sux)
}

func CreateCounterConveyor(storage MemStorage[int64]) http.Handler {
	return internal.Conveyor(CounterHandlerCreator(storage), CheckInputMiddleware)
}

func CreateGaugeConveyor(storage MemStorage[float64]) http.Handler {
	return internal.Conveyor(GaugeHandlerCreator(storage), CheckInputMiddleware)
}

func GetUrlRegexp() *regexp.Regexp {
	return regexp.MustCompile("^/[a-zA-Z][a-zA-Z0-9_]*/-?([1-9][0-9]*|0)([.][0-9]+)?$")
}

// CheckInputMiddleware
// принимаем только:
// - POST
// - Content-Type: text/plain
// - /<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
func CheckInputMiddleware(next http.Handler) http.Handler {
	pathPattern := GetUrlRegexp()
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(res, "only post methods", http.StatusMethodNotAllowed)
			return
		}
		contentType := req.Header.Get("Content-Type")
		if contentType != "" && !strings.HasPrefix(contentType, "text/plain") {
			http.Error(res, "only post methods", http.StatusUnsupportedMediaType)
			return
		}

		url := req.URL.Path

		if len(strings.Split(url, "/")) != 3 {
			http.Error(res, "wrong url", http.StatusNotFound)
			return
		}

		if !pathPattern.MatchString(url) {
			http.Error(res, "wrong url", http.StatusBadRequest)
			return
		}

		next.ServeHTTP(res, req)
	})
}

// nameValueExtractor
// разбивает сроку  /<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ> на имя и значение
// url уже валидирована!!
func nameValueExtractor(req *http.Request) (string, string) {
	url := req.URL.Path
	res := strings.Split(url, "/")[1:]
	return res[0], res[1]
}

func GaugeHandlerCreator(storage MemStorage[float64]) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		name, valueStr := nameValueExtractor(req)
		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			http.Error(res, fmt.Sprintf("wrong float64 value: %v", valueStr), http.StatusBadRequest)
			return
		}
		storage.Set(name, value)
		res.WriteHeader(http.StatusOK)
	}
}

func CounterHandlerCreator(storage MemStorage[int64]) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		name, valueStr := nameValueExtractor(req)
		value, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			http.Error(res, fmt.Sprintf("wrong int64 value: %v", valueStr), http.StatusBadRequest)
			return
		}
		storage.Add(name, value)
		res.WriteHeader(http.StatusOK)
	}
}
