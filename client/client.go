package main

import (
	"fmt"
	"github.com/sht/ambilight/engine"
	"github.com/sht/ambilight/packet"
	"log"
)

func main() {
	// Create ambilight instance
	amb, err := engine.Init("client")
	if err != nil {
		log.Fatalf("Invalid startup mode.\n")
	}
	displays, err := amb.GetDisplays()
	if err != nil {
		log.Fatalf("Could not initialize display capturers: %s\n.", err)
	}

	fmt.Println("Attempting to connect to the Ambilight server...")
	for {
		amb.Connect()
		fmt.Println("Connection established.")
		for {
			var data [1024 * 3]uint8
			offset := 0
			for _, d := range displays {
				img := engine.AcquireImage(d.Capturer, amb.Framerate)
				pix := engine.CapturePixels(img, d.Width, d.Height)
				pix = engine.FilterPixels(d, pix, d.BoundsOffset, d.BoundsSize)
				avg := engine.AveragePixels(pix, d.Leds)
				// data = append(data, ...)
				for _, b := range avg {
					data[offset] = b
				}
			}
			payload := packet.Ambilight{
				Action: 'A',
				Data:   data,
			}
			_, err = amb.Write(payload)
			if err != nil {
				// Error while writing -- connection closed by server
				break
			}
		}
		fmt.Println("Connection closed. Retrying...")
	}

}
