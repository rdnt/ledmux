package ambilight

import (
	"bufio"
	"fmt"
	"github.com/sht/ambilight/config"
	ws281x "github.com/sht/ambilight/ws281x_wrapper"
	"io"
	"net"
	"time"
)

// Ambilight represents the state and configuration of the server/client
type Ambilight struct {
	IP         string
	Port       int
	LedsCount  int
	Framerate  int
	GPIOPin    int
	Brightness int
	Displays   []*config.Display
	Mode       rune
	Ws281x     *ws281x.Engine
	reader     *bufio.Reader
}

// Init returns an ambilight object with the default values and the specified
// IP port and leds count
func Init(cfg *config.Config) *Ambilight {
	return &Ambilight{
		IP:         cfg.IP,
		Port:       cfg.Port,
		LedsCount:  cfg.LedsCount,
		GPIOPin:    cfg.GPIOPin,
		Framerate:  cfg.Framerate,
		Brightness: cfg.Brightness,
		Displays:   cfg.Displays,
		Mode:       'A',
	}
}

// Connect initializes a TCP socket connection to the destination address
// It blocks until a connection is established
func (amb Ambilight) Connect() net.Conn {
	// Format address string
	address := fmt.Sprintf("%s%s%d", amb.IP, ":", amb.Port)
	// Check for connection until one is established
	for {
		conn, err := net.Dial("tcp", address)
		if err != nil {
			time.Sleep(1 * time.Second)
		} else {
			fmt.Println("Connection established.")
			return conn
		}
	}
}

// Disconnect closes an existing TCP connection
func (amb Ambilight) Disconnect(conn net.Conn) error {
	err := conn.Close()
	return err
}

// Send sends the ambilight data to the destination
func (amb Ambilight) Send(conn net.Conn, data []byte) error {
	_, err := conn.Write(data)
	return err
}

// Listen sets up a listener and returns the connection and the listener
func (amb Ambilight) Listen() (net.Conn, net.Listener, error) {
	// Format address string
	address := fmt.Sprintf(":%d", amb.Port)
	// Establish a listener
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, nil, err
	}
	// Accept incoming connections
	conn, err := listener.Accept()
	if err != nil {
		return nil, listener, err
	}
	fmt.Printf("Incoming connection from %s\n", conn.RemoteAddr())
	return conn, listener, nil
}

// Receive reads the raw bytes from the socket, stores them in a buffer and
// then returns said buffer
func (amb Ambilight) Receive(conn net.Conn) ([]byte, error) {
	// Allocate buffer to store the incoming data
	// One more byte for the mode char
	buffer := make([]uint8, amb.LedsCount*3+1)
	// Create and store a socket reader on the struct
	// if it isn't already created
	if amb.reader == nil {
		amb.reader = bufio.NewReader(conn)
	}
	// Read the data
	_, err := io.ReadFull(amb.reader, buffer)
	if err != nil {
		return nil, err
	}
	// Return the data
	return buffer, nil
}

// DisconnectListener disconnects the listener when an error occurs while
// receiving data
func (amb Ambilight) DisconnectListener(listener net.Listener) error {
	err := listener.Close()
	return err
}
