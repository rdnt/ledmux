package config

import (
	"bytes"
	"encoding/json"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	name        string
	format      string
	DefaultMode string      `yaml:"defaultMode" json:"defaultMode"`
	CaptureType string      `yaml:"captureType" json:"captureType"`
	Server      Server      `yaml:"server" json:"server"`
	Displays    [][]Display `yaml:"displays" json:"displays"`
	Segments    []Segment   `yaml:"segments" json:"segments"`
}

type Server struct {
	Host       string `yaml:"host" json:"host"`
	Port       int    `yaml:"port" json:"port"`
	Leds       int    `yaml:"leds" json:"leds"`
	StripType  string `yaml:"stripType" json:"stripType"`
	GpioPin    int    `yaml:"gpioPin" json:"gpioPin"`
	Brightness int    `yaml:"brightness" json:"brightness"`
	BlackPoint int    `json:"blackPoint" yaml:"blackPoint"`
}

type Display struct {
	Width     int    `yaml:"width" json:"width"`
	Height    int    `yaml:"height" json:"height"`
	Left      int    `yaml:"left" json:"left"`
	Top       int    `yaml:"top" json:"top"`
	Leds      int    `yaml:"leds" json:"leds"`
	Bounds    Bounds `yaml:"bounds" json:"bounds"`
	Framerate int    `yaml:"framerate" json:"framerate"`
}

type Bounds struct {
	From   Vector2 `yaml:"from" json:"from"`
	To     Vector2 `yaml:"to" json:"to"`
	Offset int     `yaml:"-" json:"-"`
	Size   int     `yaml:"-" json:"-"`
}

type Vector2 struct {
	X int `yaml:"x" json:"x"`
	Y int `yaml:"y" json:"y"`
}

type Segment struct {
	Id   int `yaml:"id" json:"id"`
	Leds int `yaml:"leds" json:"leds"`
}

func (c *Config) Save() error {
	var b []byte

	switch c.format {
	case "json":
		var err error
		b, err = json.MarshalIndent(c, "", "  ")
		if err != nil {
			return err
		}
	case "yaml":
		var buf bytes.Buffer
		enc := yaml.NewEncoder(&buf)
		enc.SetIndent(2)
		err := enc.Encode(c)
		if err != nil {
			return err
		}

		b = buf.Bytes()
	}

	err := os.WriteFile(c.name, b, 0644)
	if err != nil {
		return err
	}

	return nil
}

func Load() (*Config, error) {
	validCfgs := map[string]string{
		"config.json": "json",
		"config.yaml": "yaml",
		"config.yml":  "yaml",
	}

	for name, format := range validCfgs {
		if _, err := os.Stat(name); err == nil {
			b, err := os.ReadFile(name)
			if err != nil {
				return nil, err
			}

			var c Config

			switch format {
			case "json":
				if err := json.Unmarshal(b, &c); err != nil {
					return nil, err
				}

				c.format = "json"
			case "yaml":
				if err := yaml.Unmarshal(b, &c); err != nil {
					return nil, err
				}

				c.format = "yaml"
			}

			c.name = name

			return &c, nil
		}
	}

	return createDefault()
}

func createDefault() (*Config, error) {
	c := Config{
		DefaultMode: "ambilight",
		CaptureType: "bitblt",
		Server: Server{
			Host:       "0.0.0.0",
			Port:       4197,
			Leds:       100,
			StripType:  "grb",
			GpioPin:    18,
			Brightness: 255,
			BlackPoint: 0,
		},
		Displays: [][]Display{
			{
				{
					Leds:      100,
					Width:     1920,
					Height:    1080,
					Left:      0,
					Top:       0,
					Framerate: 60,
					Bounds: Bounds{
						From: Vector2{X: 0, Y: 0},
						To:   Vector2{X: 0, Y: 0},
					},
				},
			},
		},
	}

	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return nil, err
	}

	err = os.WriteFile("config.json", b, 0644)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
