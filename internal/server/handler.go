package server

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/siluk00/http_protocol/internal/request"
	"github.com/siluk00/http_protocol/internal/response"
)

type Handler func(w *response.Writer, req *request.Request)
type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func ProxyHandler(w *response.Writer, req *request.Request) {
	path := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin")
	if path == "" {
		path = "/"
	}
	url := "http://httpbin.org" + path

	resp, err := http.Get(url)
	if err != nil {
		w.WriteStatusLine(500)
		body := []byte("upstream error")
		w.WriteHeaders(response.GetDefaultHeaders(len(body), "text/plain"))
		w.WriteBody(body)
		return
	}
	defer resp.Body.Close()

	headers := map[string]string{
		"Content-Type":      resp.Header.Get("Content-Type"),
		"Transfer-Encoding": "chunked",
		"Trailer":           "X-Content-SHA256, X-Content-Length",
	}
	w.WriteStatusLine(response.StatusCode(resp.StatusCode))
	w.WriteHeaders(headers)

	buf := make([]byte, 4096)
	responseBody := []byte{}
	for {
		n, err := resp.Body.Read(buf)
		responseBody = append(responseBody, buf[:n]...)
		if n > 0 {
			w.WriteChunkedBody(buf[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return //should be logged
		}
	}

	w.WriteBody([]byte("0\r\n"))
	hash := sha256.Sum256(responseBody)
	trailers := map[string]string{
		"X-Content-SHA256": fmt.Sprintf("%x", hash[:]),
		"X-Content-Length": fmt.Sprint(len(responseBody)),
	}
	w.WriteTrailers(trailers)
	w.WriteBody([]byte("\r\n"))
}
