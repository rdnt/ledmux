package engine

import (
	"fmt"
	ws281x "github.com/sht/ambilight/ws281x"
	"net"
)

// Engine asd
type Engine struct {
	IP   string `json:"ip,omitempty"`
	Port int    `json:"port"`

	LedsCount  int `json:"leds_count"`
	GPIOPin    int `json:"gpio_pin,omitempty"`
	Brightness int `json:"brightness,omitempty"`
	Framerate  int `json:"framerate,omitempty"`

	Listener   *net.TCPListener   `json:"-"`
	Clients    map[string]*Client `json:"-"`
	Connection net.Conn           `json:"-"`

	Ws *ws281x.Engine `json:"-"`

	Displays []*Display `json:"displays,omitempty"`

	Action    string `json:"-"`
	Connected bool   `json:"-"`
	Running   bool   `json:"-"`
}

// Display holds parameters like leds count and pixel offsets in the config
type Display struct {
	From Vector2 `json:"from"`
	To   Vector2 `json:"to"`
	Leds int     `json:"leds"`
}

// Vector2 is a X,Y pair of coordinates
type Vector2 struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Init returns an ambilight object with the default values and the specified
// IP port and leds count
func Init(mode string) (*Engine, error) {
	if mode == "server" {
		ws, err := ws281x.Init(18, 75, 255)
		if err != nil {
			return nil, err
		}
		return &Engine{
			Ws:         ws,
			IP:         "localhost",
			Port:       4197,
			LedsCount:  100,
			GPIOPin:    18,
			Framerate:  60,
			Brightness: 255,
			Action:     "",
			Clients:    make(map[string]*Client),
		}, nil
	} else if mode == "client" {
		return &Engine{
			IP:         "192.168.1.101",
			Port:       4197,
			LedsCount:  75,
			GPIOPin:    18,
			Framerate:  60,
			Brightness: 255,
			Action:     "",
			Displays: []*Display{
				&Display{
					From: Vector2{
						X: 300,
						Y: 1079,
					},
					To: Vector2{
						X: 1320,
						Y: 1079,
					},
					Leds: 75,
				},
			},
		}, nil
	}
	return nil, fmt.Errorf("invalid startup mode")
}

//
// // ReloadEngine asd
func (e *Engine) Reload(
	ledsCount uint16, gpioPin uint8, brightness uint8) error {
	err := e.Ws.Clear()
	if err != nil {
		return err
	}
	e.Ws.Fini()
	engine, err := ws281x.Init(int(gpioPin), int(ledsCount), int(brightness))
	if err != nil {
		return err
	}
	e.Ws = engine
	return nil
}
