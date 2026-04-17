package server

import (
	"github.com/siluk00/http_protocol/internal/request"
	"github.com/siluk00/http_protocol/internal/response"
)

type Handler func(w *response.Writer, req *request.Request)
type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}
