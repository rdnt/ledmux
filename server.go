package main

import (
    "fmt"
    "os"
    "time"
    "./ambilight"
)

func main() {

    // Create the Ambilight object
    var amb = ambilight.Init(
        "",
        4197,
        75,
    )
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
            fmt.Println("Receiving data...")
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
            Handle(data)
        }
        // Try to reconnect every second
        time.Sleep(1 * time.Second)
        fmt.Println("Re-trying...")
    }
}

func Handle(data []byte) {
    fmt.Printf("%X\n", data)
}
