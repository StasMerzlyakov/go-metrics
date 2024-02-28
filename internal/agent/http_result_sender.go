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
	var buf bytes.Buffer

	w, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		logrus.Errorf("gzip.NewWriterLevel error: %v", err)
		return err
	}
	if err := json.NewEncoder(w).Encode(metric); err != nil {
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
		SetBody(buf.Bytes()).Post(h.serverAdd + "/update/")
	if err != nil {
		logrus.Errorf("server communication error: %v", err)
	}

	if resp.StatusCode() != http.StatusOK {
		errStr := fmt.Sprintf("unexpected server http response code: %v", resp.StatusCode())
		logrus.Errorf(errStr)
		return errors.New(errStr)
	}

	/*_, err := h.client.R().
	SetHeader("Content-Type", "application/json; charset=UTF-8").
	SetBody(metric).Post(h.serverAdd + "/update/") */

	return err
}
