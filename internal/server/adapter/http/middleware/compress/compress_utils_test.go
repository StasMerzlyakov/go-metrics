package compress_test

import (
	"bytes"
	"io"
	"net/http"
	"strings"
)

type checkBodyHandler struct {
	expected []byte
}

func (ch *checkBodyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)

	if err != nil && err != io.EOF {
		http.Error(w, "read body err", http.StatusInternalServerError)
	}

	if !bytes.Equal(ch.expected, body) {
		http.Error(w, "unexpected body err", http.StatusBadRequest)
	}
}

type defaultHTMLHandle struct{}

func (defaultHTMLHandle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	io.WriteString(w, "<html><body>"+strings.Repeat("Hello, world<br>", 20)+"</body></html>")
}

type defaultTextHandle struct{}

func (defaultTextHandle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	io.WriteString(w, strings.Repeat("Hello, world\n", 20))

}

type defaultJSONHandle struct{}

func (defaultJSONHandle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, "{ "+strings.Repeat(`"msg":"Hello, world",`, 19)+`"msg":"Hello, world"`+"}")
}
