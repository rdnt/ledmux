package ambilight

import (
	"bufio"
	"fmt"
	"gopkg.in/ini.v1"
	"io"
	"log"
	"net"
	"os"
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
	Reader    *bufio.Reader
	Running   bool
	Buffer    []byte
}

// Init returns an ambilight object with the default values and the specified
// IP port and leds count
func Init() *Ambilight {
	cfg, err := ini.Load("ambilight.conf")
	if err != nil {
		createConfig()
	}
	addr := cfg.Section("").Key("ip").String()
	ip := net.ParseIP(addr)
	if ip.To4() == nil {
		log.Fatal(addr, ": not a valid IPv4 address.")
	}
	port := getIntKey(cfg, "port", 1024, 65535)
	count := getIntKey(cfg, "leds_count", 1, 65535)
	fps := getIntKey(cfg, "framerate", 1, 144)
	return &Ambilight{
		IP:        addr,
		Port:      port,
		Count:     count,
		Framerate: fps,
		Running:   false,
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

func createConfig() {
	cfg := ini.Empty()
	sec, err := cfg.GetSection("")
	if err != nil {
		log.Fatal("Failed to create the new configuration.")
	}
	sec.NewKey("ip", "127.0.0.1")
	sec.NewKey("port", "4197")
	sec.NewKey("leds_count", "64")
	sec.NewKey("framerate", "60")
	sec.NewKey("pwm_pin", "18")
	sec.NewKey("brightness", "255")

	cfg.SaveTo("ambilight.conf")
	log.Print("Default configuration file created. Please edit and relaunch the client.")
	os.Exit(0)
}

func getIntKey(cfg *ini.File, key string, min int, max int) int {
	str := cfg.Section("").Key(key).String()
	value, err := strconv.Atoi(str)
	if err != nil || value < min || value > max {
		log.Fatalf("%s: %s out of range (%d - %d)", str, key, min, max)
	}
	return value
}
