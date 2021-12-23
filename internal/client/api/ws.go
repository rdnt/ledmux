package api

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  0, //1024
	WriteBufferSize: 0,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	EnableCompression: false,
}

func (h *Hub) AcceptWebsocketConnection(w http.ResponseWriter, req *http.Request) {
	wsconn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	wsconn.EnableWriteCompression(false)

	client := h.Add(wsconn)

	for {
		b, err := client.Receive()
		if err != nil {
			h.Remove(client.id)
			return
		}

		// trigger the event
		h.evts <- Event{
			ClientId: client.id,
			Data:     b,
		}
	}
}
