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

			ch <- b[:n]
		}
	}()

	return ch
}

func (s *server) Close() error {
	return s.conn.Close()
}

func NewServer(address string) (*server, error)  {
	addr, err := net.ResolveUDPAddr("mockserver", address)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("mockserver", addr)
	if err != nil {
		return nil, err
	}

	return &server{
		conn: conn,
	}, nil
}
