package agent_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/StasMerzlyakov/go-metrics/internal/agent"
	"github.com/stretchr/testify/require"
)

func TestHash256Header_No_Key(t *testing.T) {

	mux := http.NewServeMux()
	mux.HandleFunc("/updates/", func(w http.ResponseWriter, r *http.Request) {
		require.Empty(t, r.Header.Get("HashSHA256"))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	sender := agent.NewHTTPResultSender(srv.URL, "")

	value := 1.
	sender.SendMetrics(context.Background(), []agent.Metrics{
		{
			ID:    "HeapReleased",
			MType: agent.GaugeType,
			Value: &value,
		},
	})
}

func TestHash256Header_With_Key(t *testing.T) {

	mux := http.NewServeMux()
	mux.HandleFunc("/updates/", func(w http.ResponseWriter, r *http.Request) {
		require.NotEmpty(t, r.Header.Get("HashSHA256"))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	sender := agent.NewHTTPResultSender(srv.URL, "test_key")

	value := 1.
	sender.SendMetrics(context.Background(), []agent.Metrics{
		{
			ID:    "HeapReleased",
			MType: agent.GaugeType,
			Value: &value,
		},
	})
}
