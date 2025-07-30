package request

import (
	"fmt"
	"io"
	"slices"
	"strings"
)

type Request struct {
	RequestLine RequestLine
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
	done
)

const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0
	request := Request{
		State: initialized,
	}

	for request.State != done {
		if readToIndex >= bufferSize {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		n, err := reader.Read(buf[readToIndex:])
		readToIndex += n

		if err != nil {
			request.State = done
			break
		}

		bytesReaden, err := request.parse(buf)

		if err != nil {
			return nil, err
		}

		newBuf := make([]byte, len(buf))
		copy(newBuf, buf[bytesReaden:])
		buf = newBuf
		readToIndex -= bytesReaden

	}

	return &request, nil
}

func getMethodList() []string {
	return []string{"GET", "POST"}
}

func parseRequestLine(line string) (*RequestLine, int, error) {
	if !strings.Contains(line, "\r\n") {
		return nil, 0, nil
	}

	lines := strings.Split(line, "\r\n")
	parts := strings.Split(lines[0], " ")
	if len(parts) != 3 {
		return nil, 0, fmt.Errorf("bad line")
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
	}, len(lines[0]), nil
}

func (r *Request) parse(data []byte) (int, error) {

	if r.State == initialized {
		requestLine, bytesReaden, err := parseRequestLine(string(data))
		if err != nil {
			return 0, err
		}

		if bytesReaden == 0 {
			return 0, nil
		}

		r.State = done
		r.RequestLine = *requestLine
		return bytesReaden, nil
	}

	if r.State == done {
		return 0, fmt.Errorf("there's nothing to be readen from ")
	}

	return 0, fmt.Errorf("unknown state")

}
