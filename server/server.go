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
	var led_count, brightness, port, pin int
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
	led_count, err = strconv.Atoi(args[0])
	if err != nil || led_count < 1 || led_count > 65535 {
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
		led_count,
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

func Handle(amb *ambilight.Ambilight, data []byte, reset chan struct{}) {
	// Parse operation mode and remove the byte from the data
	mode, data := parseMode(data)
	// If mode is valid, execute it
	//fmt.Println(time.Now().Format("02/01/2016 15:04:05 " + string(mode)))
	Render(amb, mode, data, reset)
}

// Mode that reproduces the LED data that are sent
func Render(amb *ambilight.Ambilight, mode rune, data []byte, reset chan struct{}) {
	// For each of the LEDs in the received data
	if amb.Running == false {
		amb.Running = true
	} else {
		reset <- struct{}{}
	}
	//Reset(amb);
	//}
	switch mode {
	case 'R':
		go Rainbow(amb, data, reset)
	case 'A':
		go Ambilight(amb, data, reset)
	default:
		go Nullify(amb, data, reset)
	}
}

func Reset2(amb *ambilight.Ambilight) {
	// Clear the leds
	ws2811.Clear()
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

func SetLedColor(led int, r int, g int, b int) {
	color := uint32(0xFF)<<24 | uint32(g)<<16 | uint32(r)<<8 | uint32(b)
	ws2811.SetLed(led, color)
}
