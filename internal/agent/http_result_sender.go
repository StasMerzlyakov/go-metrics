package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"net/http"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

func NewHTTPResultSender(serverAdd string, hash256Key string) *httpResultSender {
	if !strings.HasPrefix(serverAdd, "http") {
		serverAdd = "http://" + serverAdd
	}
	serverAdd = strings.TrimSuffix(serverAdd, "/")
	return &httpResultSender{
		serverAdd:  serverAdd,
		client:     resty.New(),
		batchSize:  5,
		hash256Key: hash256Key,
	}
}

type httpResultSender struct {
	serverAdd  string
	client     *resty.Client
	batchSize  int
	hash256Key string
}

func (h *httpResultSender) SendMetrics(ctx context.Context, metrics []Metrics) error {
	logrus.Infof("SendMetrics start")
	for i := 0; i*h.batchSize < len(metrics); i++ {
		end := (i + 1) * h.batchSize
		if (i+1)*h.batchSize > len(metrics) {
			end = len(metrics)
		}

		if err := h.store(ctx, metrics[i*h.batchSize:end]); err != nil {
			return err
		}
	}
	return nil
}

func (h *httpResultSender) store(ctx context.Context, metrics []Metrics) error {
	var buf bytes.Buffer

	var wc io.WriteCloser

	wc, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		logrus.Errorf("gzip.NewWriterLevel error: %v", err)
		return err
	}

	if h.hash256Key != "" {
		hashWriter := &hashWriter{
			hasher:      hmac.New(sha256.New, []byte(h.hash256Key)),
			WriteCloser: wc,
		}
		wc = hashWriter
	}

	if err := json.NewEncoder(wc).Encode(metrics); err != nil {
		logrus.Errorf("json encode error: %v", err)
		return err
	}

	err = wc.Close()
	if err != nil {
		logrus.Errorf("gzip close error: %v", err)
	}

	request := h.client.R().
		SetHeader("Content-Type", "application/json; charset=UTF-8").
		SetHeader("Content-Encoding", "gzip")

	if hsr, ok := wc.(*hashWriter); ok {
		hexValue := hex.EncodeToString(hsr.Sum())
		request.SetHeader("HashSHA256", hexValue)
	}

	resp, err := request.
		SetBody(buf.Bytes()).
		SetContext(ctx).Post(h.serverAdd + "/updates/")

	if err != nil {
		logrus.Errorf("server communication error: %v", err)
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		errStr := fmt.Sprintf("unexpected server http response code: %v", resp.StatusCode())
		logrus.Errorf(errStr)
		return errors.New(errStr)
	}

	return err
}

type hashWriter struct {
	hasher hash.Hash
	io.WriteCloser
}

func (hw *hashWriter) Sum() []byte {
	return hw.hasher.Sum(nil)
}

func (hw *hashWriter) Write(p []byte) (n int, err error) {
	_, err = hw.hasher.Write(p)
	if err != nil {
		return 0, err
	}

	return hw.WriteCloser.Write(p)
}
