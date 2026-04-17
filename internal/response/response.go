package response

import (
	"fmt"
	"io"

	"github.com/siluk00/http_protocol/internal/headers"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	_, err := w.Write([]byte(fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, mapStatusLine(statusCode))))
	return err
}

func mapStatusLine(statusCode StatusCode) string {
	switch statusCode {
	case 200:
		return "OK"
	case 400:
		return "Bad Request"
	case 500:
		return "Internal Server Error"
	default:
		return ""
	}
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")
	return h
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for k, v := range headers {
		if _, err := w.Write([]byte(fmt.Sprintf("%s: %s\r\n", k, v))); err != nil {
			return err
		}
	}
	if _, err := w.Write([]byte("\r\n")); err != nil {
		return err
	}
	return nil
}
