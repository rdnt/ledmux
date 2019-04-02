package main

import (
    "fmt"
    "os"
    "time"
    "strconv"
    "./ambilight"
	"github.com/jgarff/rpi_ws281x/golang/ws2811"
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
    Reset()
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
                Reset()
                break
            }
            // Data handler function
            Handle(data, amb.Count)
        }
        // Try to reconnect every second
        time.Sleep(1 * time.Second)
    }
}

func parseMode(data []byte) (string, []byte) {
    // Split the mode character and the led data and return them
    return string(data[0]), data[1:]
}

func Handle(data []byte, count int) {
    // Parse operation mode and remove the byte from the data
    mode, data := parseMode(data)
    // If mode is valid, execute it
    if mode == "R" {
        Render(data, count)
    } else {
        fmt.Println("Invalid mode supplied.")
    }
}

// Mode that reproduces the LED data that are sent
func Render(data []byte, count int) {
    // Initialize variables
    var r, g, b uint8
    var color uint32
    // For each of the LEDs in the received data
    for i := 0; i < count; i++ {
        // Calcu;ate offset from start
        offset := (i + 3 * 8) % count
        // Parse color data for current LED
        r = uint8(data[i * 3])
        g = uint8(data[i * 3 + 1])
        b = uint8(data[i * 3 + 2])
        // FGRB
        color = uint32(0xFF)<<24 | uint32(g)<<16 | uint32(r) <<8 | uint32(b)
        // Set the current LED's color
        ws2811.SetLed(offset, color)
    }
    // Render the leds
    err := ws2811.Render()
    if err != nil {
        fmt.Println("Error while rendering the LEDs.")
    }
}

func Reset() {
    // Clear the leds
    ws2811.Clear()
	err := ws2811.Render()
    // Error handling in case we can't clear the leds
	if err != nil {
		fmt.Println("Error while resetting the LEDs.")
        os.Exit(1)
	}
}
