package engine

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/satori/go.uuid"
	"github.com/sht/ambilight/packet"
)

// Client asda
type Client struct {
	ID   string
	Conn *net.TCPConn
}

// Connect a
func (amb *Engine) Connect() {
	// Check for connection every second until one is established
	for {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", amb.IP, amb.Port), 1*time.Second)
		if err == nil {
			amb.Connection = conn
			return
		}
	}
}

func (amb *Engine) Write(payload interface{}) (int, error) {
	var buffer bytes.Buffer
	err := binary.Write(&buffer, binary.BigEndian, payload)
	if err != nil {
		return 0, err
	}
	return amb.Connection.Write(buffer.Bytes())
}

// Listen sets up a listener and returns the connection and the listener
func (amb *Engine) Listen() error {
	// Validate tcp address
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%d", amb.Port))
	if err != nil {
		return err
	}
	// Create a listener
	lstn, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}
	amb.Listener = lstn
	return nil
}

// AcceptConnection starts a new tcp connection for each new client
func (amb *Engine) AcceptConnection(handler func(*Engine, []byte)) error {
	conn, err := amb.Listener.AcceptTCP()
	if err != nil {
		return err
	}
	fmt.Printf("Incoming connection from %s\n", conn.RemoteAddr())
	// id, err := uuid.NewV4()
	// if err != nil {
	// 	fmt.Printf("Error while generating client UUID: %s.\n", err)
	// 	return nil
	// }
	// cid := id.String()
	cid := uuid.NewV4().String()
	// if err != nil {
	// 	fmt.Printf("Error while generating client UUID: %s.\n", err)
	// 	return nil
	// }
	// cid := id.String()
	c := &Client{
		ID:   cid,
		Conn: conn,
	}
	amb.Clients[cid] = c
	go amb.HandleConnection(c, handler)
	return nil
}

// DisconnectClient asd
func (amb *Engine) DisconnectClient(id string) {
	fmt.Println("Client", id, "disconnected")
	delete(amb.Clients, id)
}

// HandleConnection asd
func (amb *Engine) HandleConnection(c *Client, handler func(*Engine, []byte)) {
	// Initialize a reader for this connection
	reader := bufio.NewReader(c.Conn)
	// Create buffer for this specific connection, max size should be 3 * 1024
	// (max 1024 leds) + 1 byte for the action
	buffer := make([]byte, 1+1024*3)
	for {
		// While we are receiving data, peek into the data and check what kind
		// of packet it is based on its action
		size := 0
		action, err := reader.Peek(1)
		if err != nil {
			// Error while reading, connection closed by client -- cleanup
			amb.DisconnectClient(c.ID)
			return
		}
		switch string(action) {
		case "U":
			// Get size of the packet
			size = packet.Update{}.Size()
			// Read up to the size
			_, err := io.ReadAtLeast(reader, buffer, size)
			if err != nil {
				// Reading failed; skip packet
				fmt.Println(err)
				continue
			}
		case "C", "R":
			// Cancel packet is 1 byte only (the action)
			size = 1
			_, err := io.ReadAtLeast(reader, buffer, size)
			if err != nil {
				// Reading failed; skip packet
				fmt.Println(err)
				continue
			}
		case "A":
			size = packet.Ambilight{}.Size()
			_, err := io.ReadAtLeast(reader, buffer, size)
			if err != nil {
				// Reading failed; skip packet
				fmt.Println(err)
				continue
			}
		default:
			// Invalid action, force disconnect of the client
			fmt.Printf("Invalid action: %s\n", action)
			amb.DisconnectClient(c.ID)
			return
		}
		// Pass the packet data to the handler function
		handler(amb, buffer)
	}
}

// NewEngine asd
func (amb *Engine) NewEngine() {
	// amb := ambilight.Init(cfg)

}
