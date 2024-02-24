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

func (h *httpResultSender) SendMetrics(metrics []Metric) error {
	for _, metric := range metrics {
		var httpType string
		if metric.Type == GaugeType {
			httpType = "gauge"
		} else {
			httpType = "counter"
		}
		err := h.store(httpType, metric.Name, metric.Value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *httpResultSender) store(metricType string, metricName string, value string) error {
	_, err := h.client.R().
		SetHeader("Content-Type", "text/plain; charset=UTF-8").
		SetPathParams(map[string]string{
			"metricType": metricType,
			"metricName": metricName,
			"value":      value,
		}).Post(h.serverAdd + "/update/{metricType}/{metricName}/{value}")

	return err
}
