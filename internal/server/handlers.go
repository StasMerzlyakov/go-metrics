package server

import (
	"github.com/StasMerzlyakov/go-metrics/internal"
	"github.com/StasMerzlyakov/go-metrics/internal/storage"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

func CreateCounterHandler(storage storage.MetricsStorage[int64]) http.Handler {
	return internal.Conveyor(CounterHandlerCreator(storage), CheckInputMiddleware)
}

func CreateGaugeHandler(storage storage.MetricsStorage[float64]) http.Handler {
	return internal.Conveyor(GaugeHandlerCreator(storage), CheckInputMiddleware)
}

// CheckInputMiddleware
// принимаем только:
// - POST
// - Content-Type: text/plain
// - /<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
func CheckInputMiddleware(next http.Handler) http.Handler {
	pathPattern := getURLRegexp()
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

func GaugeHandlerCreator(storage storage.MetricsStorage[float64]) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		name, value, err := extractFloat64(req)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		storage.Set(name, value)
		res.WriteHeader(http.StatusOK)
	}
}

func CounterHandlerCreator(storage storage.MetricsStorage[int64]) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		name, value, err := extractInt64(req)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		storage.Add(name, value)
		res.WriteHeader(http.StatusOK)
	}
}

// getURLRegexp
// возвращает regexp для проверки url вида  /<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
func getURLRegexp() *regexp.Regexp {
	return regexp.MustCompile("^/[a-zA-Z][a-zA-Z0-9_]*/-?([1-9][0-9]*|0)([.][0-9]+)?$")
}

// nameValueExtractor
// разбивает сроку  /<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ> на имя и значение
// url уже валидирована!!
func nameValueExtractor(req *http.Request) (string, string) {
	url := req.URL.Path
	res := strings.Split(url, "/")[1:]
	return res[0], res[1]
}

func extractFloat64(req *http.Request) (string, float64, error) {
	name, valueStr := nameValueExtractor(req)
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return name, -1, err
	}
	return name, value, nil
}

func extractInt64(req *http.Request) (string, int64, error) {
	name, valueStr := nameValueExtractor(req)
	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		return name, -1, err
	}
	return name, value, nil
}
