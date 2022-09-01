package tcp

import (
	"bytes"
	"fmt"
	"io"
	"net"
)

type Server interface {
	Receive() chan []byte
}

type server struct {
	conn *net.TCPListener
}

func (s *server) Receive() chan []byte {
	ch := make(chan []byte)

	go func() {
		for {
			conn, err := s.conn.Accept()
			if err != nil {
				fmt.Println("Socket acceptance error: " + err.Error())
				continue
			}

			fmt.Println("socket accepted")
			go func(conn net.Conn) {
				defer conn.Close()

				result := bytes.NewBuffer(nil)
				var buf [1024]byte
				for {
					n, err := conn.Read(buf[0:])
					result.Write(buf[0:n])
					if err != nil {
						if err == io.EOF {
							continue
						} else {
							fmt.Println("read err:", err)
							break
						}
					} else {
						b := result.Bytes()
						result.Reset()

						go func() {
							ch <- b
						}()
					}
				}

				//for {
				//	b := make([]byte, 65535)
				//	n, err := conn.Read(b)
				//	if err != nil {
				//		fmt.Println(err)
				//		return
				//	}
				//
				//	if n == 0 {
				//		fmt.Print("i")
				//		continue
				//	}
				//
				//	go func() {
				//		ch <- b[:n]
				//	}()
				//}
			}(conn)
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
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &server{
		conn: conn,
	}, nil
}
