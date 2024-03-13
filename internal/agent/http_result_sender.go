package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
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
		batchSize: 5,
	}
}

type httpResultSender struct {
	serverAdd string
	client    *resty.Client
	batchSize int
}

func (h *httpResultSender) SendMetrics(metrics []Metrics) error {
	for i := 0; i*h.batchSize < len(metrics); i++ {
		end := (i + 1) * h.batchSize
		if (i+1)*h.batchSize > len(metrics) {
			end = len(metrics)
		}
		if err := h.store(metrics[i*h.batchSize : end]); err != nil {
			return err
		}
	}
	return nil
}

func (h *httpResultSender) store(metrics []Metrics) error {
	var buf bytes.Buffer

	w, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		logrus.Errorf("gzip.NewWriterLevel error: %v", err)
		return err
	}
	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		logrus.Errorf("json encode error: %v", err)
		return err
	}

	err = w.Close()
	if err != nil {
		logrus.Errorf("gzip close error: %v", err)
	}

	resp, err := h.client.R().
		SetHeader("Content-Type", "application/json; charset=UTF-8").
		SetHeader("Content-Encoding", "gzip").
		SetBody(buf.Bytes()).Post(h.serverAdd + "/updates/")
	if err != nil {
		logrus.Errorf("server communication error: %v", err)
	}

	if resp.StatusCode() != http.StatusOK {
		errStr := fmt.Sprintf("unexpected server http response code: %v", resp.StatusCode())
		logrus.Errorf(errStr)
		return errors.New(errStr)
	}

	return err
}
