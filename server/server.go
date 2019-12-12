package main

import (
	"fmt"
	"github.com/sht/ambilight/ambilight"
	"github.com/sht/ambilight/config"
	ws281x "github.com/sht/ambilight/ws281x_wrapper"
	"log"
	"time"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error while loading configuration: %s.\n", err)
	}
	// Create the Ambilight object
	amb := ambilight.Init(cfg)
	// Initialize the leds
	amb.Ws281x, err = ws281x.Init(amb.GPIOPin, amb.LedsCount, amb.Brightness)
	if err != nil {
		log.Fatalf("Error while initializing ws281x library: %s.\n", err)
	}
	// Clear the leds and finish gracefully uppon exit
	defer func() {
		_ = amb.Ws281x.Clear()
		amb.Ws281x.Fini()
	}()
	// Reset the LEDs before we start, just in case
	Reset(amb)
	fmt.Printf("Ambilight server listening on port %d...\n", amb.Port)
	// Try to re-establish socket connection
	for {
		// Establish connection to the local socket
		conn, listener, err := amb.Listen()
		if err != nil {
			time.Sleep(1 * time.Second)
			// Attempt to connect every second
			continue
		}
		// Stop channel is used to signal any running goroutine to stop
		// rendering and return
		stop := make(chan struct{})
		// Initialize mode character
		var mode rune
		// Receive data indefinitely
		for {
			data, err := amb.Receive(conn)
			if err != nil {
				// Disconnect the listener
				err := amb.DisconnectListener(listener)
				if err != nil {
					log.Fatalf("Connection could not be closed: %s.\n", err)
				}
				// Connection lost, reset the leds
				close(stop)
				// Reset leds if ambilight client is disconnected
				if mode == 'A' {
					Reset(amb)
				}
				// Break the loop, we will attempt to get new connections now
				break
			}
			// Split mode from data
			mode, data = ParseMode(data)
			if mode != amb.Mode {
				amb.Mode = mode
				close(stop)
				stop = make(chan struct{})
			}
			// Render the leds based on the given mode and data
			go Render(amb, mode, data, stop)
		}
		// Try to reconnect every second
		time.Sleep(1 * time.Second)
	}
}

// ParseMode splits the mode and data apart and returns them
func ParseMode(data []byte) (rune, []byte) {
	// Split the mode character and the led data and return them
	return rune(data[0]), data[1:]
}

// Render calls the appropriate rendering function. If none matches, it resets
// the leds
func Render(amb *ambilight.Ambilight, mode rune, data []byte, stop chan struct{}) {
	// Call appropriate routine based on the supplied mode
	switch mode {
	case 'R':
		go Rainbow(amb, data, stop)
	case 'A':
		go Ambilight(amb, data, stop)
	default:
		// Reset leds if mode is invalid
		Reset(amb)
	}
}

// Reset sets each led's color to black (switches it off)
func Reset(amb *ambilight.Ambilight) {
	// Ignoring errors
	_ = amb.Ws281x.Clear()
}

// Rainbow loops a smooth gradient color swipe across the led strip
func Rainbow(amb *ambilight.Ambilight, data []byte, stop chan struct{}) {
	for {
		var r, g, b int
		// Loop color brightness 3 times
		for i := 0; i < 256*3; i++ {
			// For each of the leds
			for j := 0; j < amb.LedsCount; j++ {
				if (i+j)%768 < 256 {
					// transition from red to green
					r = (255 - (i+j)%256)
					g = (i + j) % 256
					b = 0
				} else if (i+j)%768 < 512 {
					// transition from green to blue
					r = 0
					g = (255 - (i+j)%256)
					b = (i + j) % 256
				} else {
					// transition from blue to red
					r = (i + j) % 256
					g = 0
					b = (255 - (i+j)%256)
				}
				select {
				default:
					// Not need to check for error
					_ = amb.Ws281x.SetLedColor(j, uint8(r), uint8(g), uint8(b))
				case <-stop:
					// Stop executing if signaled from main process
					return
				}
			}
			// Render the leds, ignoring errors
			_ = amb.Ws281x.Render()
			time.Sleep(16 * time.Millisecond)
		}
	}
}

// Ambilight simply sets each led's color based on the received data
func Ambilight(amb *ambilight.Ambilight, data []byte, stop chan struct{}) {
	select {
	default:
		// Initialize variables
		var r, g, b uint8
		for i := 0; i < amb.LedsCount; i++ {
			// Parse color data for current LED
			r = uint8(data[i*3])
			g = uint8(data[i*3+1])
			b = uint8(data[i*3+2])
			// Set the current LED's color
			// Not need to check for error
			_ = amb.Ws281x.SetLedColor(i, r, g, b)
		}
		// Render the leds, ignoring errors
		_ = amb.Ws281x.Render()
	case <-stop:
		// Stop executing if signaled from main process
		return

	}
}
