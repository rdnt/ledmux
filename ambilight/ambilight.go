package ambilight

import (
	"../config"
	"bufio"
	"fmt"
	"gopkg.in/ini.v1"
	"io"
	"log"
	"net"
	"strconv"
	"time"
)

// Ambilight represents the state and configuration of the server/client
type Ambilight struct {
	conn      net.Conn
	IP        string
	Port      int
	Count     int
	Framerate int
	Displays  []*config.Display
	Reader    *bufio.Reader
	Running   bool
	Buffer    []byte
}

// Init returns an ambilight object with the default values and the specified
// IP port and leds count
func Init(cfg *config.Config) *Ambilight {
	return &Ambilight{
		IP:        cfg.IP,
		Port:      cfg.Port,
		Count:     cfg.LedsCount,
		Framerate: cfg.Framerate,
		Running:   false,
		Displays:  cfg.Displays,
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
	fmt.Printf("Incoming connection from %s\n", conn.RemoteAddr())
	//mb.Reader = bufio.NewReader(conn)
	return conn, listener, nil
}

// Receive reads the raw bytes from the socket, stores them in a buffer and
// then returns said buffer
func (amb Ambilight) Receive(conn net.Conn) ([]byte, error) {
	// Allocate buffer to store the incoming data
	// One more byte for the mode char
	buffer := make([]uint8, amb.Count*3+1)
	// Create and store a socket reader on the struct
	// if it isn't already created
	if amb.Reader == nil {
		amb.Reader = bufio.NewReader(conn)
	}
	// Read the data
	_, err := io.ReadFull(amb.Reader, buffer)
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

func getIntKey(cfg *ini.File, key string, min int, max int) int {
	str := cfg.Section("").Key(key).String()
	value, err := strconv.Atoi(str)
	if err != nil || value < min || value > max {
		log.Fatalf("%s: %s out of range (%d - %d)", str, key, min, max)
	}
	return value
}
