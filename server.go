package main

import (
    "fmt"
    "os"
    "time"
    "./ambilight"
)

const IP = ""
const port = 4197
const count = 75

func main() {

    // Create the Ambilight object
    var amb = ambilight.Init(
        IP,
        port,
        count,
    )

    for {
        fmt.Println("Listening for Ambilight connection requests on port 4197...")
        // Connect to remote server
        conn, listener, err := amb.Listen()
        if err != nil {
            fmt.Println(err)
            time.Sleep(1 * time.Second)
            continue
        }
        // Receive data
        for {
            fmt.Println("Receiving data...")
            //err := amb.Send(conn, data)
            data, err := amb.Receive(conn)
            if err != nil {
                fmt.Println("Failed to receive data.")
                err := amb.DisconnectListener(listener)
                if err != nil {
                    fmt.Println("Connection could not be closed.")
                    os.Exit(2)
                }
                fmt.Println("Connection closed.")
                break
            }

            fmt.Printf("%X\n", data)



        }
        // Try to reconnect every second
        time.Sleep(1 * time.Second)
    }
    //
    // fmt.Println("Listening for Ambilight data on port 4197...")
    //
    // // listen on all interfaces
    // listener, err := net.Listen("tcp", ":4197")
    // if err != nil {
    //     fmt.Println(err)
    // }
    //
    // // accept connection on port
    // conn, err := listener.Accept()
    // if err != nil {
    //     fmt.Println(err)
    // }
    //
    // // run loop forever (or until ctrl-c)
    // for {
    //     // Create a buffer the size of the led count * 3 (for RGB bytes)
    //     buffer := make([]uint8, count * 3)
    //     input := bufio.NewReader(conn)
    //     _, err := io.ReadFull(input, buffer)
    //     if err != nil {
    //         fmt.Println(err)
    //     }
    //     // output message received
    //     fmt.Printf("%X\n", buffer)
    // }
}
