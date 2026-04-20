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

type state int

const (
	writingHeaders state = iota
	writingBody
)

type Writer struct {
	Writer     io.Writer
	writeState state
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	_, err := w.Writer.Write([]byte(fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, mapStatusLine(statusCode))))
	if err != nil {
		w.writeState = writingHeaders
	}
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

func GetDefaultHeaders(contentLen int, contentType string) headers.Headers {

	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", contentType)
	return h
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.writeState != writingHeaders {
		return fmt.Errorf("wrong order of writing")
	}

	for k, v := range headers {
		if _, err := w.Writer.Write([]byte(fmt.Sprintf("%s: %s\r\n", k, v))); err != nil {
			return err
		}
	}
	if _, err := w.Writer.Write([]byte("\r\n")); err != nil {
		return err
	}
	w.writeState = writingBody
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.writeState != writingBody {
		return 0, fmt.Errorf("wrong order of writing")
	}
	return w.Writer.Write(p)
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	header := []byte(fmt.Sprintf("%x\r\n", len(p)))
	n := len(header)
	if _, err := w.WriteBody(header); err != nil {
		return 0, err
	}

	written, err := w.WriteBody(p)
	if err != nil {
		return n, err
	}
	n += written

	written, err = w.WriteBody([]byte("\r\n"))
	if err != nil {
		return n, err
	}

	return n + written, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	return w.WriteBody([]byte("0\r\n\r\n"))
}

func (w *Writer) WriteTrailers(header headers.Headers) error {
	for h, v := range header {
		_, err := w.WriteBody([]byte(fmt.Sprintf("%s: %s\r\n", h, v)))
		if err != nil {
			return err
		}
	}
	return nil
}
