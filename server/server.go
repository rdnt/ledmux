package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"

	"github.com/sht/ambilight/effects"
	"github.com/sht/ambilight/engine"
	"github.com/sht/ambilight/packet"
	ws281x "github.com/sht/ambilight/ws281x"
)

var ws *ws281x.Engine

func main() {
	// Create ambilight instance
	amb, err := engine.Init("server")
	if err != nil {
		log.Fatalf("Invalid startup mode.\n")
	}

	// Create a TCP listener on the specified port
	err = amb.Listen()
	if err != nil {
		log.Fatalf("Could not bind to port %d: %s.\n", amb.Port, err)
	}
	fmt.Printf("Ambilight server listening on port %d...\n", amb.Port)
	for {
		err = amb.AcceptConnection(Handler)
		if err != nil {
			fmt.Printf("Failed to accept connection: %s.\n", err)
			continue
		}
	}
}

// Handler asd
func Handler(amb *engine.Engine, payload []byte) {
	action := string(payload[0])
	// fmt.Printf("%s: %d\n", action, payload)
	if action != amb.Action {
		// A different action provided, stop existing workers and set it
		amb.Ws.Stop()
		amb.Action = action
	} else if action != "A" {
		// If action is not ambilight and it's the same as the old one, return
		return
	}
	// Create a binary reader for the payload
	reader := bytes.NewReader(payload)
	switch action {
	case "U":
		// Unmarshal update packet
		var data packet.Update
		err := binary.Read(reader, binary.BigEndian, &data)
		if err != nil {
			fmt.Printf("Could not unmarshal request payload: %s.\n", err)
		}
		fmt.Printf("%+v\n", data)
		err = amb.Reload(
			data.LedsCount, data.GPIOPin, data.Brightness,
		)
		// ws, err := ReloadEngine(
		// 	int(data.LedsCount), int(data.GPIOPin), int(data.Brightness),
		// )
		// if err != nil {
		// 	fmt.Printf("Failed to reload engine: %s.\n", err)
		// }
		err = amb.SaveConfig()
		if err != nil {
			fmt.Printf("Failed to reload engine: %s.\n", err)
		}
		fmt.Printf("Engine reload complete.\n")
	case "R":
		fmt.Println("Rainbow!")
		go effects.Rainbow(amb.Ws, payload)
	case "C":
		fmt.Println("Cancel!")
	case "A":
		go effects.Ambilight(amb.Ws, payload[3:])
	}
}
