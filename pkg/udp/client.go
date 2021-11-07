package udp

import (
	"net"
)

type Client interface {
	Send(b []byte) error
}

type client struct {
	conn *net.UDPConn
}

func (c *client) Send(b []byte) error {
	_, err := c.conn.Write(b)
	if err != nil {
       return err
    }

	return nil
}

func (c *client) Close() error {
	return c.conn.Close()
}

func NewClient(address string) (*client, error)  {
	addr, err := net.ResolveUDPAddr("mockserver", address)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("mockserver", nil, addr)
	if err != nil {
		return nil, err
	}

	return &client{
		conn: conn,
	}, nil
}
