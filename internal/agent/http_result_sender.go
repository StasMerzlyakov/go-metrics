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
	"net"
	"net/http"
	"strings"

	"github.com/StasMerzlyakov/go-metrics/internal/config"
	"github.com/StasMerzlyakov/go-metrics/internal/keygen"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/singleflight"
)

func NewHTTPResultSender(conf *config.AgentConfiguration) *httpResultSender {

	srvAddr := conf.ServerAddr

	if !strings.HasPrefix(srvAddr, "http") {
		srvAddr = "http://" + srvAddr
	}
	srvAddr = strings.TrimSuffix(srvAddr, "/")

	sender := &httpResultSender{
		serverAdd:  srvAddr,
		client:     resty.New(),
		hash256Key: conf.Key,
		getIpGroup: &singleflight.Group{},
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
	agentIP    string
	getIpGroup *singleflight.Group
}

// ParserServerAddr return net.IP by http address
func ParserServerAddr(serverHttpAddr string) string {
	// serverHttpAddr examples:
	//    localhost
	//    127.0.0.1
	// 	  127.0.0.1:8081
	//    http://127.0.0.1/
	//    https://127.0.0.1:8081/
	serverAdd := strings.TrimSuffix(serverHttpAddr, "/")

	defaultPort := "80" // default http port

	if strings.Contains(serverAdd, "://") {
		addrs := strings.Split(serverAdd, "://")
		if addrs[0] == "https" {
			defaultPort = "443"
		}
		serverAdd = addrs[1]
	}

	if !strings.Contains(serverAdd, ":") {
		serverAdd = serverAdd + ":" + defaultPort
	}

	return serverAdd
}

// GetDialIP return preferred IP outbound for server communication
func GetDialIP(ctx context.Context, serverAddr string) (net.IP, error) {
	serverAddr = ParserServerAddr(serverAddr)

	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", serverAddr)

	if err != nil {
		return nil, err
	}
	defer conn.Close()
	lAddr := conn.LocalAddr()

	localAddress := lAddr.(*net.TCPAddr)

	return localAddress.IP, nil
}

func (h *httpResultSender) GetIP(ctx context.Context) (string, error) {
	if h.agentIP == "" {
		netIPInterface, err, _ := h.getIpGroup.Do(h.serverAdd, func() (interface{}, error) {
			ip, err := GetDialIP(ctx, h.serverAdd)
			if err != nil {
				return "", err
			}
			return ip, err

		})
		if err != nil {
			return "", err
		}
		h.agentIP = netIPInterface.(net.IP).String()
	}
	return h.agentIP, nil
}

func (h *httpResultSender) Stop() {
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
		return err
	}

	clientIP, err := h.GetIP(ctx)
	if err != nil {
		logrus.Errorf("get clinet ip error: %v", err)
		return err
	}

	request := h.client.R().
		SetHeader("Content-Type", "application/json; charset=UTF-8").
		SetHeader("Content-Encoding", "gzip").
		SetHeader("X-Real-IP", clientIP)

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
