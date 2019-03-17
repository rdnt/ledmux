package main

import (
    "fmt"
    "os"
    "time"
    "./ambilight"
	"github.com/jgarff/rpi_ws281x/golang/ws2811"
)

func main() {
    // Get all arguments except for program
    args := os.Args[1:]
    // Make sure we get exactly 4 arguments
    if len(args) != 4 {
        fmt.Println("Usage: ./server [port] [pin] [led_count] [brightness]")
        return
    }
    // Validate controller port is in allowed range (1024 - 65535)
    port, err := strconv.ParseUint(args[0], 10, 16)
    if err != nil || port < 1024 {
        fmt.Println(args[0], ": Port out of range. (1024 - 65535)")
        return
    }
    // Validate PWM pin number to send the led data to
    pin, err := strconv.ParseUint(args[1], 10, 6)
    if err != nil || pin != 12 || pin != 13 || pin != 18 || pin != 19 {
        fmt.Println(args[1], ": Invalid hardware PWM pin: (12 / 13 / 18 / 19)")
        return
    }
    // Validate leds count
    led_count, err := strconv.ParseUint(args[2], 10, 16)
    if err != nil || led_count == 0 {
        fmt.Println(args[2], ": Invalid LED count. (1 - 65535)")
        return
    }
    // Validate brightness is in allowed range (0 - 255)
    brightness, err := strconv.ParseUint(args[3], 10, 8)
    if err != nil {
        fmt.Println("Brightness: 0 - 255", brightness)
        return
    }
    // Create the Ambilight object
    var amb = ambilight.Init(
        "",
        port,
        led_count,
    )
	defer ws2811.Fini()
	err := ws2811.Init(18, amb.Count, 255)
	if err != nil {
		fmt.Println(err)
		return
	}

    fmt.Println("Initializing server...")
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
                fmt.Println("Failed to receive data.")
                // Disconnect the listener
                err := amb.DisconnectListener(listener)
                if err != nil {
                    fmt.Println("Connection could not be closed.")
                    os.Exit(2)
                }
                fmt.Println("Connection closed.")
                break
            }
            // Data handler function
            Handle(data, amb.Count)
        }
        // Try to reconnect every second
        time.Sleep(1 * time.Second)
        fmt.Println("Re-trying...")
    }
}

func Handle(data []byte, count int) {
    //fmt.Printf("%X\n", data)



	var r, g, b uint8
	var color uint32

	for i := 0; i < count; i++ {

		offset := (i + 24) % count

		r = uint8(data[i * 3])
		g = uint8(data[i * 3 + 1])
		b = uint8(data[i * 3 + 2]) // GRB
		color = uint32(0xFF)<<24 | uint32(g)<<16 | uint32(r) <<8 | uint32(b)
		ws2811.SetLed(offset, color)
	}
	err := ws2811.Render()
	if err != nil {
		fmt.Println(err)
	}
}
