package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"net/http"

	"github.com/StasMerzlyakov/go-metrics/internal/config"
	"github.com/StasMerzlyakov/go-metrics/internal/keygen"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

func NewHTTPResultSender(conf *config.AgentConfiguration) *httpResultSender {

	sender := &httpResultSender{
		serverAdd:  conf.ServerAddr,
		client:     resty.New(),
		hash256Key: conf.Key,
	}

	if conf.CryptoKey != "" {
		pubKey, err := keygen.ReadPubKey(conf.CryptoKey)
		if err != nil {
			panic(err)
		}
		sender.pubKey = pubKey
	}

	return sender
}

type httpResultSender struct {
	serverAdd  string
	client     *resty.Client
	hash256Key string
	pubKey     *rsa.PublicKey
}

func (h *httpResultSender) SendMetrics(ctx context.Context, metrics []Metrics) error {
	var buf bytes.Buffer

	var wc io.WriteCloser

	wc, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		logrus.Errorf("gzip.NewWriterLevel error: %v", err)
		return err
	}

	if h.pubKey != nil {
		wc = &encryptedWriter{
			key:         h.pubKey,
			WriteCloser: wc,
		}
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

type encryptedWriter struct {
	key *rsa.PublicKey
	io.WriteCloser
	buf bytes.Buffer
}

func (ew *encryptedWriter) Write(p []byte) (int, error) {
	return ew.buf.Write(p)
}

func (ew *encryptedWriter) Close() error {
	encrypted, err := keygen.EncryptWithPublicKey(ew.buf.Bytes(), ew.key)
	if err != nil {
		return err
	}
	_, err = ew.WriteCloser.Write(encrypted)

	if err != nil {
		return err
	}

	return ew.WriteCloser.Close()
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
