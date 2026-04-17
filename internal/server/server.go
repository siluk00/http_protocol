package server

import (
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"

	"github.com/siluk00/http_protocol/internal/request"
	"github.com/siluk00/http_protocol/internal/response"
)

type Server struct {
	errChan  chan error
	listener net.Listener
	isClosed atomic.Bool
	wg       sync.WaitGroup
	handler  Handler
}

func Serve(handler Handler, port int) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	if err != nil {
		return nil, fmt.Errorf("couln1t produce listener: %v", err)
	}

	server := Server{
		listener: listener,
		errChan:  make(chan error, 100),
		handler:  handler,
	}

	server.wg.Add(1)
	go server.listen()
	go server.manageErrors()

	return &server, nil
}

func (s *Server) manageErrors() {
	for err := range s.errChan {
		log.Printf("Server Error: %v\n", err)
	}
}

func (s *Server) Close() error {
	s.isClosed.Swap(true)

	if err := s.listener.Close(); err != nil {
		return err
	}

	s.wg.Wait()
	close(s.errChan)
	return nil
}

func (s *Server) listen() {
	defer s.wg.Done()
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.isClosed.Load() {
				return
			}
			s.errChan <- fmt.Errorf("listener accept error: %v", err)
			continue
		}
		s.wg.Add(1)
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	defer s.wg.Done()

	writer := &response.Writer{
		Writer: conn,
	}
	req, err := request.RequestFromReader(conn)
	if err != nil {
		writer.WriteStatusLine(400)
		writer.WriteHeaders(response.GetDefaultHeaders(0, "text/plain"))
		return
	}

	s.handler(writer, req)
}
