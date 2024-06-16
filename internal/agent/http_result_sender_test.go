package agent_test

import (
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/StasMerzlyakov/go-metrics/internal/agent"
	"github.com/StasMerzlyakov/go-metrics/internal/config"
	"github.com/StasMerzlyakov/go-metrics/internal/keygen"
	"github.com/stretchr/testify/require"
)

func TestHash256Header_No_Key(t *testing.T) {

	mux := http.NewServeMux()
	mux.HandleFunc("/updates/", func(w http.ResponseWriter, r *http.Request) {
		require.Empty(t, r.Header.Get("HashSHA256"))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	clntConf := config.AgentConfiguration{
		ServerAddr: srv.URL,
		Key:        "",
	}

	sender := agent.NewHTTPResultSender(&clntConf)

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

	clntConf := config.AgentConfiguration{
		ServerAddr: srv.URL,
		Key:        "test_key",
	}

	sender := agent.NewHTTPResultSender(&clntConf)

	value := 1.
	sender.SendMetrics(context.Background(), []agent.Metrics{
		{
			ID:    "HeapReleased",
			MType: agent.GaugeType,
			Value: &value,
		},
	})
}

func TestHash256Header_Encrypt(t *testing.T) {

	tempDir := os.TempDir()

	pubKeyFile, err := os.CreateTemp(tempDir, "*")
	require.NoError(t, err)
	defer os.Remove(pubKeyFile.Name())

	privKeyFile, err := os.CreateTemp(tempDir, "*")
	require.NoError(t, err)
	defer os.Remove(privKeyFile.Name())

	err = keygen.Create(pubKeyFile, privKeyFile)
	require.NoError(t, err)

	privKey, err := keygen.ReadPrivKey(privKeyFile.Name())
	require.NoError(t, err)

	mux := http.NewServeMux()
	mux.HandleFunc("/updates/", func(w http.ResponseWriter, r *http.Request) {

		zr, err := gzip.NewReader(r.Body)
		require.NoError(t, err)

		encrypted, err := io.ReadAll(zr)

		if err != nil && err != io.EOF {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		defer r.Body.Close()

		_, err = keygen.DecryptWithPrivateKey(encrypted, privKey)
		require.NoError(t, err)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	clntConf := config.AgentConfiguration{
		ServerAddr: srv.URL,
		Key:        "",
		CryptoKey:  pubKeyFile.Name(),
	}

	sender := agent.NewHTTPResultSender(&clntConf)

	value := 1.
	sender.SendMetrics(context.Background(), []agent.Metrics{
		{
			ID:    "HeapReleased",
			MType: agent.GaugeType,
			Value: &value,
		},
	})
}
