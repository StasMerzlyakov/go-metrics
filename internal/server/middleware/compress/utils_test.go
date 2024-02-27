package compress_test

import (
	"io"
	"net/http"
	"strings"
)

type defaultHtmlHandle struct{}

func (defaultHtmlHandle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	io.WriteString(w, "<html><body>"+strings.Repeat("Hello, world<br>", 20)+"</body></html>")
}

type defaultTextHandle struct{}

func (defaultTextHandle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	io.WriteString(w, strings.Repeat("Hello, world\n", 20))

}

type defaultJsonHandle struct{}

func (defaultJsonHandle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, "{ "+strings.Repeat(`"msg":"Hello, world",`, 19)+`"msg":"Hello, world"`+"}")
}
