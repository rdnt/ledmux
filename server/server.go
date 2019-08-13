package main

import (
	"../ambilight"
	"../ws2811"
	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {
	// Get all arguments except for program
	args := os.Args[1:]
	var ledCount, brightness, port, pin int
	var err error
	// Make sure we get 2 to 4 arguments
	if len(args) > 4 || len(args) < 2 {
		fmt.Println("Usage: ./server [led_count] [brightness] (pin) (port)")
		return
	}
	if len(args) == 2 {
		pin = 18
		port = 4197
	} else if len(args) == 3 {
		// Validate PWM pin number to send the led data to
		pin, err = strconv.Atoi(args[2])
		if err != nil || !(pin == 12 || pin == 13 || pin == 18 || pin == 19) {
			fmt.Println(args[2], ": Invalid hardware PWM pin: (12 / 13 / 18 / 19)")
			return
		}
		port = 4197
	} else if len(args) == 4 {
		// Validate PWM pin number to send the led data to
		pin, err = strconv.Atoi(args[2])
		if err != nil || !(pin == 12 || pin == 13 || pin == 18 || pin == 19) {
			fmt.Println(args[2], ": Invalid hardware PWM pin: (12 / 13 / 18 / 19)")
			return
		}
		// Validate controller port is in allowed range (1024 - 65535)
		port, err = strconv.Atoi(args[3])
		if err != nil || port < 1024 || port > 65535 {
			fmt.Println(args[3], ": Port out of range. (1024 - 65535)")
			return
		}
	}

	// Validate leds count
	ledCount, err = strconv.Atoi(args[0])
	if err != nil || ledCount < 1 || ledCount > 65535 {
		fmt.Println(args[0], ": Invalid LED count. (1 - 65535)")
		return
	}
	// Validate brightness is in allowed range (0 - 255)
	brightness, err = strconv.Atoi(args[1])
	if err != nil || brightness < 1 || brightness > 255 {
		fmt.Println(args[1], ": Invalid Brightness. (1 - 255)", brightness)
		return
	}

	// Create the Ambilight object
	var amb = ambilight.Init(
		"",
		port,
		ledCount,
	)

	// Initialize the leds
	err = ws2811.Init(pin, amb.Count, brightness)
	if err != nil {
		fmt.Println(err)
		return
	}
	// Clear the leds and finish gracefully uppon exit
	defer ws2811.Clear()
	defer ws2811.Fini()
	// Reset the LEDs just in case
	Reset(amb)
	// Try to re-establish socket connection
	for {
		// Establish connection to the local socket
		conn, listener, err := amb.Listen()
		if err != nil {
			fmt.Println(err)
			time.Sleep(1 * time.Second)
			// Keep trying to connect
			continue
		}
		// Receive data indefinitely
		reset := make(chan struct{})
		for {
			//fmt.Println("Receiving data...")
			data, err := amb.Receive(conn)
			if err != nil {
				// Disconnect the listener
				err := amb.DisconnectListener(listener)
				if err != nil {
					fmt.Println("Connection could not be closed.")
					os.Exit(2)
				}
				fmt.Println("Connection closed.")
				// Connection lost, reset the leds
				reset <- struct{}{}
				amb.Running = false
				Reset(amb)
				break
			}
			// Data handler function
			Handle(amb, data, reset)
		}
		// Try to reconnect every second
		time.Sleep(1 * time.Second)
	}
}

func parseMode(data []byte) (rune, []byte) {
	// Split the mode character and the led data and return them
	return rune(data[0]), data[1:]
}

// Handle decodes the request data and renders the leds based on the mode
// that is supplied
func Handle(amb *ambilight.Ambilight, data []byte, reset chan struct{}) {
	// Parse operation mode and remove the byte from the data
	mode, data := parseMode(data)
	// If mode is valid, execute it
	//fmt.Println(time.Now().Format("02/01/2016 15:04:05 " + string(mode)))
	Render(amb, mode, data, reset)
}

// Render resets the leds and calls the appropriate rendering function
func Render(amb *ambilight.Ambilight, mode rune, data []byte, reset chan struct{}) {
	// Stop current mode and reset the leds' state
	if amb.Running == false {
		amb.Running = true
	} else {
		reset <- struct{}{}
	}
	// Call appripriate routine based on the supplied mode
	switch mode {
	case 'R':
		go Rainbow(amb, data, reset)
	case 'A':
		go Ambilight(amb, data, reset)
	default:
		go Nullify(amb, data, reset)
	}
}

// Nullify will reset the state of the leds when an invalid mode is supplied
func Nullify(amb *ambilight.Ambilight, data []byte, reset chan struct{}) {
	for {
		select {
		default:
			Reset(amb)
		case <-reset:
			return
		}
	}
}

// Reset sets each led's color to black (switches it off)
func Reset(amb *ambilight.Ambilight) {
	for i := 0; i < amb.Count; i++ {
		SetLedColor(i, 0, 0, 0)
	}
	err := ws2811.Render()
	// Error handling in case we can't clear the leds
	if err != nil {
		fmt.Println("Error while resetting the LEDs.")
		os.Exit(1)
	}
}

// Rainbow loops a smooth gradient color swipe across the led strip
func Rainbow(amb *ambilight.Ambilight, data []byte, reset chan struct{}) {
	for {
		var r, g, b int
		for i := 0; i < 256*3; i++ {
			for j := 0; j < amb.Count; j++ {
				if (i+j)%768 < 256 {
					r = (255 - (i+j)%256)
					g = (i + j) % 256
					b = 0
				} else if (i+j)%768 < 512 {
					r = 0
					g = (255 - (i+j)%256)
					b = (i + j) % 256
				} else {
					r = (i + j) % 256
					g = 0
					b = (255 - (i+j)%256)
				}
				select {
				default:
					SetLedColor(j, r, g, b)
				case <-reset:
					return
				}
			}
			// Render the leds
			err := ws2811.Render()
			if err != nil {
				fmt.Println("Error while rendering the LEDs.")
			}
			time.Sleep(16 * time.Millisecond)
		}
	}
}

// Ambilight simply sets each led's color based on the received data
func Ambilight(amb *ambilight.Ambilight, data []byte, reset chan struct{}) {
	for {
		select {
		default:
			// Initialize variables
			var r, g, b uint8
			var color uint32
			for i := 0; i < amb.Count; i++ {
				// Calculate offset from start
				offset := (i + 3*8) % amb.Count
				// Parse color data for current LED
				r = uint8(data[i*3])
				g = uint8(data[i*3+1])
				b = uint8(data[i*3+2])
				// AGRB
				color = uint32(0xFF)<<24 | uint32(g)<<16 | uint32(r)<<8 | uint32(b)
				// Set the current LED's color
				ws2811.SetLed(offset, color)
			}
			// Render the leds
			err := ws2811.Render()
			if err != nil {
				fmt.Println("Error while rendering the LEDs.")
			}
		case <-reset:
			return
		}
	}
}

// SetLedColor changes the color the led in the specified index
func SetLedColor(led int, r int, g int, b int) {
	color := uint32(0xFF)<<24 | uint32(g)<<16 | uint32(r)<<8 | uint32(b)
	ws2811.SetLed(led, color)
}
