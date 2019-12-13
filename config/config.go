package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
)

// Config represents the ambilight configuration file which is stored as json
type Config struct {
	Port       int `json:"port"`
	LedsCount  int `json:"leds_count"`
	GPIOPin    int `json:"gpio_pin"`
	Brightness int `json:"brightness"`

	IP        string     `json:"ip,omitempty"`
	Framerate int        `json:"framerate,omitempty"`
	Displays  []*Display `json:"displays,omitempty"`
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

// Load returns the configuration struct or an error if loading/creation failed
func Load() (*Config, error) {
	cfg := new(Config)
	f, err := os.OpenFile("ambilight.conf", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	enc, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(enc, cfg)
	if err != nil {
		fmt.Println(err)
		// Error while unmarshalling json, create new config file
		cfg, err = createConfig(f)
		if err != nil {
			return nil, err
		}
	}
	ip := net.ParseIP(cfg.IP)
	ipv4 := ip.To4()
	if ipv4 == nil ||
		cfg.Port < 1024 || cfg.Port > 65535 ||
		cfg.LedsCount <= 0 || cfg.LedsCount > 65535 ||
		cfg.Framerate < 1 || cfg.Framerate > 144 ||
		cfg.GPIOPin <= 0 || cfg.GPIOPin > 40 ||
		cfg.Brightness <= 0 || cfg.Brightness > 255 {
		return nil, fmt.Errorf("configuration file is corrupted")
	}
	if len(cfg.Displays) == 0 || len(cfg.Displays) > 4 {
		return nil, fmt.Errorf("maximum 4 displays supported")
	}
	total := 0
	for i, d := range cfg.Displays {
		if d.From.X < 0 || d.From.Y < 0 || d.To.X < 0 || d.To.Y < 0 {
			return nil, fmt.Errorf("invalid pixel offsets for display %d", i+1)
		}
		if d.Leds <= 0 {
			return nil, fmt.Errorf("configuration file is corrupted")
		}
		total += d.Leds
	}
	if total != cfg.LedsCount {
		return nil, fmt.Errorf("led counts do not add up")
	}
	return cfg, nil
}

func createConfig(f *os.File) (*Config, error) {
	cfg := Config{
		IP:         "192.168.1.1",
		Port:       4197,
		LedsCount:  100,
		Framerate:  60,
		GPIOPin:    18,
		Brightness: 255,
		Displays: []*Display{
			&Display{
				From: Vector2{0, 0},
				To:   Vector2{0, 0},
				Leds: 100,
			},
		},
	}
	enc, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, err
	}
	err = f.Truncate(0)
	if err != nil {
		return nil, err
	}
	_, err = f.Seek(0, 0)
	if err != nil {
		return nil, err
	}
	_, err = f.Write(enc)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Save updates the saved configuration file with in-memory values
func (cfg *Config) Save(f *os.File) error {
	f, err := os.OpenFile("ambilight.conf", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	enc, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	err = f.Truncate(0)
	if err != nil {
		return err
	}
	_, err = f.Seek(0, 0)
	if err != nil {
		return err
	}
	_, err = f.Write(enc)
	if err != nil {
		return err
	}
	return nil
}
