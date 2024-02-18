package agent

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/go-resty/resty/v2"
)

func NewHTTPResultSender(serverAdd string) ResultSender {
	if !strings.HasPrefix(serverAdd, "http") {
		serverAdd = "http://" + serverAdd
	}
	serverAdd = strings.TrimSuffix(serverAdd, "/")
	return &httpResultSender{
		serverAdd: serverAdd,
	}
}

type httpResultSender struct {
	serverAdd string
	client    *resty.Client
	sm        sync.Mutex
}

func (h *httpResultSender) initIfNecessary() {
	h.sm.Lock()
	defer h.sm.Unlock()
	if h.client == nil {
		h.client = resty.New().
			// Иногда возникает ошибка EOF или http: server closed idle connection; добавим Retry
			SetRetryCount(3)
	}
}

func (h *httpResultSender) store(metricType string, metricName string, value string) error {
	h.initIfNecessary()
	_, err := h.client.R().
		SetHeader("Content-Type", "text/plain; charset=UTF-8").
		SetPathParams(map[string]string{
			"metricType": metricType,
			"metricName": metricName,
			"value":      value,
		}).Post(h.serverAdd + "/update/{metricType}/{metricName}/{value}")

	if err != nil {
		fmt.Printf("%v\n", errors.Unwrap(err))
	}

	return err
}

func (h *httpResultSender) SendGauge(name string, value float64) error {
	return h.store("gauge", name, fmt.Sprintf("%v", value))
}

func (h *httpResultSender) SendCounter(name string, value int64) error {
	return h.store("counter", name, fmt.Sprintf("%v", value))
}
