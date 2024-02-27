package agent

import (
	"strings"

	"github.com/go-resty/resty/v2"
)

func NewHTTPResultSender(serverAdd string) *httpResultSender {
	if !strings.HasPrefix(serverAdd, "http") {
		serverAdd = "http://" + serverAdd
	}
	serverAdd = strings.TrimSuffix(serverAdd, "/")
	return &httpResultSender{
		serverAdd: serverAdd,
		client: resty.New().
			// Иногда возникает ошибка EOF или http: server closed idle connection; добавим Retry
			SetRetryCount(3),
	}
}

type httpResultSender struct {
	serverAdd string
	client    *resty.Client
}

func (h *httpResultSender) SendMetrics(metrics []Metrics) error {
	for _, metric := range metrics {
		err := h.store(metric)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *httpResultSender) store(metric Metrics) error {
	_, err := h.client.R().
		SetHeader("Content-Type", "application/json; charset=UTF-8").
		SetBody(metric).Post(h.serverAdd + "/update/")
	return err
}
