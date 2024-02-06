package agent

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"strings"
	"sync"
)

func NewHTTPResultSender(serverAdd string) ResultSender {
	if !strings.HasPrefix(serverAdd, "http") {
		serverAdd = "http://" + serverAdd
	}
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
		h.client = resty.New()
		h.serverAdd = strings.TrimSuffix(h.serverAdd, "/")
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
	return err
}

func (h *httpResultSender) SendGauge(name string, value float64) error {
	return h.store("gauge", name, fmt.Sprintf("%v", value))
}

func (h *httpResultSender) SendCounter(name string, value int64) error {
	return h.store("counter", name, fmt.Sprintf("%v", value))
}
