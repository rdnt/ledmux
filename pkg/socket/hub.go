package socket

import (
	"errors"
	"sync"
)

var (
	ErrNotFound = errors.New("not found")
)

type Hub struct {
	sync.Mutex
	conns map[string]Client
}

func (h *Hub) Add(c Client) error {
	h.Lock()
	defer h.Unlock()

	h.conns[c.Id()] = c

	return nil
}

func (h *Hub) Client(id string) (Client, error) {
	h.Lock()
	defer h.Unlock()

	c, ok := h.conns[id]
	if !ok {
		return nil, ErrNotFound
	}

	return c, nil
}

func (h *Hub) Clients() ([]Client, error) {
	h.Lock()
	defer h.Unlock()

	clients := []Client{}
	for _, c := range h.conns {
		clients = append(clients, c)
	}

	return clients, nil
}

func (h *Hub) DeleteClient(id string) error {
	h.Lock()
	defer h.Unlock()

	_, ok := h.conns[id]
	if !ok {
		return ErrNotFound
	}
	delete(h.conns, id)

	return nil
}

func newHub() *Hub {
	return &Hub{
		conns: make(map[string]Client, 0),
	}
}
