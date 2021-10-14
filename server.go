package fsync

import (
	"context"
	"io"
	"log"
	"net"
	"sync"
)

type Server struct {
	listener net.Listener
	halt     func()
	wg       sync.WaitGroup
}

func NewServer(addr string) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	s := &Server{
		halt: cancel,
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil
	}

	s.listener = l
	s.wg.Add(1)
	go s.serve(ctx)

	return s
}

func (s *Server) Stop() {
	s.halt()
	s.listener.Close()
	s.wg.Wait()
}

func (s *Server) serve(ctx context.Context) {
	defer s.wg.Done()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				log.Println("accept error", err)
				continue
			}
		}

		s.wg.Add(1)
		go func() {
			defer s.wg.Done()

			s.handleConection(conn)
		}()

	}
}

func (s *Server) handleConection(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 2048)
	for {
		n, err := conn.Read(buf)
		if err != nil && err != io.EOF {
			log.Println("read error", err)
			return
		}
		if n == 0 {
			return
		}
		log.Printf("received from %v: %s", conn.RemoteAddr(), string(buf[:n]))
	}
}
