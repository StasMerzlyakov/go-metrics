package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/http/handler"
	"github.com/StasMerzlyakov/go-metrics/internal/server/domain"
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestAdminOperation_Ping(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := NewMockAdminApp(ctrl)

	m.EXPECT().Ping(gomock.Any()).Return(nil).Times(1)

	r := chi.NewRouter()

	log := logger()
	domain.SetMainLogger(log)
	handler.AddAdminOperations(r, m)

	srv := httptest.NewServer(r)
	defer srv.Close()

	req := resty.New().R()
	req.Method = http.MethodGet

	req.URL = srv.URL + "/ping"
	req.Header.Add("Content-Type", handler.TextPlain)
	_, err := req.Send()
	require.Nil(t, err)
}
