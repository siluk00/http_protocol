package request

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/siluk00/http_protocol/internal/headers"
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	State       state
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type state int

const (
	initialized state = iota
	initializingHeaders
	done
)

const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readenIndex := 0
	request := Request{
		State:   initialized,
		Headers: headers.NewHeaders(),
	}

	for request.State != done {
		if readenIndex == len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		n, err := reader.Read(buf[readenIndex:])

		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		readenIndex += n
		for {
			bytesReaden, err := request.parse(buf[:readenIndex])

			if err != nil {
				return nil, err
			}

			if bytesReaden == 0 {
				break
			}

			//newBuf := make([]byte, len(buf))
			//copy(newBuf, buf[bytesReaden:])
			//buf = newBuf
			copy(buf, buf[bytesReaden:readenIndex])
			readenIndex -= bytesReaden

			if request.State == done {
				return &request, nil
			}
		}

	}

	return &request, nil
}

func getMethodList() []string {
	return []string{"GET", "POST"}
}

// returns request line or nothing and bytes readen
func parseRequestLine(lines string) (*RequestLine, int, error) {
	index := strings.Index(lines, "\r\n")
	if index == -1 {
		return nil, 0, nil
	}

	line := lines[:index]
	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return nil, 0, fmt.Errorf("bad requeest line: expected 3 parts")
	}

	if !slices.Contains(getMethodList(), parts[0]) {
		return nil, 0, fmt.Errorf("method unrecognized")
	}

	if parts[2] != "HTTP/1.1" {
		return nil, 0, fmt.Errorf("unsupported protocol")
	}

	return &RequestLine{
		HttpVersion:   "1.1",
		RequestTarget: parts[1],
		Method:        parts[0],
	}, len(line) + 2, nil
}

// wrapper to request line, mounts the struct r with request line and return bytes readen
func (r *Request) parse(data []byte) (int, error) {

	switch r.State {
	case initialized:
		requestLine, bytesReaden, err := parseRequestLine(string(data))
		if err != nil {
			return 0, err
		}

		if bytesReaden == 0 {
			return 0, nil
		}

		r.State = initializingHeaders
		r.RequestLine = *requestLine
		return bytesReaden, nil
	case initializingHeaders:
		n, finished, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}

		if n == 0 {
			return 0, nil
		}

		if finished {
			r.State = done
		}

		return n, nil
	case done:
		return 0, fmt.Errorf("there's nothing to be readen from ")
	default:
		return 0, fmt.Errorf("unknown state")
	}
}
