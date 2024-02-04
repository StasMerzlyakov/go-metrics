package agent

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

func NewHTTPResultSender(serverAdd string, contentType string) ResultSender {
	return &httpResultSender{
		serverAdd:   serverAdd,
		contentType: contentType,
		urlPattern:  "/update/%v/%v/%v",
	}
}

type httpResultSender struct {
	serverAdd   string
	contentType string
	urlPattern  string
	client      *http.Client
	sm          sync.Mutex
}

func (h *httpResultSender) initIfNecessary() error {
	if h.client == nil {
		h.sm.Lock()
		defer h.sm.Unlock()
		if h.client == nil {
			h.client = &http.Client{}
			h.serverAdd = strings.TrimSuffix(h.serverAdd, "/")
		}
	}
	return nil
}

func (h *httpResultSender) store(url string) error {
	if err := h.initIfNecessary(); err != nil {
		return err
	}
	fullURL := h.serverAdd + url
	res, err := h.client.Post(fullURL, h.contentType, nil)
	if err != nil {
		fmt.Printf("server interation error: %v\n", err.Error()) // log error
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		resBody, err := io.ReadAll(res.Body)
		// log error
		if err == nil {
			fmt.Printf("server error: \n    status: %v\n    content: %v\n", res.StatusCode, string(resBody))
		} else {
			fmt.Printf("server error: \n    status: %v\n    content read error: %v\n", res.StatusCode, err.Error())
		}
		return ErrServerInteraction
	}
	return nil
}

func (h *httpResultSender) SendGauge(name string, value float64) error {
	url := fmt.Sprintf(h.urlPattern, "gauge", name, value)
	return h.store(url)
}

func (h *httpResultSender) SendCounter(name string, value int64) error {
	url := fmt.Sprintf(h.urlPattern, "counter", name, value)
	return h.store(url)
}
