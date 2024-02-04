package server

import (
	"context"
	"github.com/StasMerzlyakov/go-metrics/internal"
	"github.com/StasMerzlyakov/go-metrics/internal/storage"
	"net/http"
)

type Configuration struct {
}

func CreateServer(ctx context.Context, config Configuration) error {
	counterStorage := storage.NewMemoryInt64Storage()
	gaugeStorage := storage.NewMemoryFloat64Storage()
	sux := http.NewServeMux()
	sux.HandleFunc("/", internal.BadRequestHandler)
	sux.Handle("/update/gauge/", http.StripPrefix("/update/gauge", CreateGaugeHandler(gaugeStorage)))
	sux.Handle("/update/counter/", http.StripPrefix("/update/counter", CreateCounterHandler(counterStorage)))
	return http.ListenAndServe(`:8080`, sux)
}
