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

	b1 := make([]byte, 99*4)
	b2 := make([]byte, 99*4)

	msgs := server.Receive()

	fmt.Println("start")
	
	for msg := range msgs {
		var e events.AmbilightEvent

		err := msgpack.Unmarshal(msg, &e)
		if err != nil {
			panic(err)
		}

		if e.SegmentId == 0 {
			b1 = e.Data
		} else if e.SegmentId == 1 {
			b2 = e.Data
		}

		b := append(b1, b2...)

		fmt.Print("\r")
		for i := 0; i < len(b); i += 4 {
			color.RGB(b[i], b[i+1], b[i+2], true).Print(" ")
		}
	}
}
