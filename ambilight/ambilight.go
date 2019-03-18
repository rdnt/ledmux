package ambilight

import (
    "net"
    "fmt"
    "time"
    "bufio"
    "io"
)

type Ambilight struct {
    conn net.Conn
    IP string
    Port int
    Count int
}

func Init(IP string, port int, count int) *Ambilight {
    amb := new(Ambilight)
    amb.IP = IP
    amb.Port = port
    amb.Count = count
    return amb
}

func (amb Ambilight) Connect() net.Conn {
    // Format address string
    address := fmt.Sprintf("%s%s%d", amb.IP, ":", amb.Port)
    // Check for connection until one is established
    for {
        conn, err := net.Dial("tcp", address)
        if err != nil {
            fmt.Println("Connection failed, retrying...")
            time.Sleep(1 * time.Second)
        } else {
            fmt.Println("Connection established.")
            return conn
        }
    }
}

func (amb Ambilight) Disconnect(conn net.Conn) error {
    err := conn.Close()
    return err
}


func (amb Ambilight) Send(conn net.Conn, data []byte) error {
    _, err := conn.Write(data)
    return err
}

func (amb Ambilight) Listen() (net.Conn, net.Listener, error) {
    // Format address string
    address := fmt.Sprintf("%s%s%d", amb.IP, ":", amb.Port)
    // Establish a listener
    listener, err := net.Listen("tcp", address)
    if err != nil {
        return nil, nil, err
    }
    fmt.Printf("Listening on port %d...\n", amb.Port)
    // Accept incoming connections
    conn, err := listener.Accept()
    if err != nil {
        return nil, listener, err
    }
    return conn, listener, nil
}

func (amb Ambilight) Receive(conn net.Conn) ([]byte, error) {
    // Allocate buffer to store the incoming data
    buffer := make([]uint8, amb.Count * 3)
    // Read the data
    input := bufio.NewReader(conn)
    _, err := io.ReadFull(input, buffer)
    if err != nil {
        return nil, err
    }
    // Return the data
    return buffer, nil
}

func (amb Ambilight) DisconnectListener(listener net.Listener) error {
    err := listener.Close()
    return err
}
