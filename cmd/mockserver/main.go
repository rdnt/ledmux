package main

import (
	"fmt"
	"github.com/gookit/color"
	"github.com/vmihailenco/msgpack/v5"
	"ledctl3/internal/pkg/events"
	"ledctl3/pkg/udp"
)

func main() {
	server, err := udp.NewServer(":4197")
	if err != nil {
		panic(err)
	}

	msgs := server.Receive()
	for msg := range msgs {
		var e events.AmbilightEvent

		err := msgpack.Unmarshal(msg, &e)
		if err != nil {
			panic(err)
		}

		if e.GroupId == 1 {
			continue
		}

		fmt.Print("\r")
		for i := 0; i < len(e.Data); i += 4 {
			color.RGB(e.Data[i], e.Data[i+1], e.Data[i+2], true).Print("  ")
		}
	}
}
