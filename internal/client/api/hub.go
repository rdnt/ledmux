package api

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/lithammer/shortuuid/v3"
)

type ConnectionManager interface {
	Send(id string, b []byte)
	Broadcast(b []byte)
	Events() chan Event
}

type Hub struct {
	mux   sync.Mutex
	conns map[string]*Client
	evts  chan Event
}

func (h *Hub) Events() chan Event {
	return h.evts
}

func (h *Hub) Add(conn *websocket.Conn) *Client {
	id := shortuuid.New()
	defer fmt.Println("Client", id, "connected")

	h.mux.Lock()
	defer h.mux.Unlock()

	c := NewClient(conn)
	h.conns[c.id] = c

	return c
}

func (h *Hub) Remove(id string) {
	defer fmt.Println("Client", id, "disconnected")

	h.mux.Lock()
	defer h.mux.Unlock()

	delete(h.conns, id)
}

func (h *Hub) Broadcast(b []byte) {
	h.mux.Lock()
	defer h.mux.Unlock()

	for id, conn := range h.conns {
		err := conn.Send(b)
		if err != nil {
			delete(h.conns, id)
		}
	}
}

func (h *Hub) Send(id string, b []byte) {
	h.mux.Lock()
	defer h.mux.Unlock()

	c, ok := h.conns[id]
	if !ok {
		return
	}

	err := c.Send(b)
	if err != nil {
		delete(h.conns, id)
	}
}

type Event struct {
	ClientId string
	Data     []byte
}

func New() *Hub {
	return &Hub{
		conns: make(map[string]*Client),
		evts:  make(chan Event),
	}
}
