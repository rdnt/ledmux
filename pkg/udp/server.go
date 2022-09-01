package udp

import (
	"fmt"
	"net"
)

type Server interface {
	Receive() chan []byte
}

type server struct {
	conn *net.UDPConn
}

func (s *server) Receive() chan []byte {
	ch := make(chan []byte)

	go func() {
		for {
			b := make([]byte, 65535)
			n, _, err := s.conn.ReadFromUDP(b)
			if err != nil {
				fmt.Println(err)
				return
			}

			if n == 0 {
				fmt.Print("i")
				continue
			}

			ch <- b[:n]
		}
	}()

	return ch
}

func (s *server) Close() error {
	if s.conn == nil {
		return nil
	}

	err := s.conn.Close()
	s.conn = nil
	return err
}

func NewServer(address string) (*server, error) {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}

	return &server{
		conn: conn,
	}, nil
}
