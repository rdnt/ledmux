package api

import (
	"sync"

	"github.com/gorilla/websocket"
	"github.com/lithammer/shortuuid/v3"
)

type Client struct {
	id   string
	mux  sync.Mutex
	conn *websocket.Conn
}

func (c *Client) Send(b []byte) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	return c.conn.WriteMessage(websocket.TextMessage, b)
}

func (c *Client) Receive() ([]byte, error) {
	_, b, err := c.conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	return b, nil
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		id:   shortuuid.New(),
		conn: conn,
	}
}
