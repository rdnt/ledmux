package client

import (
	"fmt"
	ws281x "github.com/rpi-ws281x/rpi-ws281x-go"
	"ledctl3/internal/client/config"
	"ledctl3/internal/client/controller"
	"ledctl3/internal/client/controller/ambilight/capturer/bitblt"
	"ledctl3/internal/client/controller/ambilight/capturer/dxgi"
	"ledctl3/internal/client/controller/ambilight/capturer/scrap"
	"net"
)

type CapturerType string

const (
	DXGI   CapturerType = "dxgi"
	BitBlt CapturerType = "bitblt"
	Scrap  CapturerType = "scrap"
)

var capturerTypes = map[string]CapturerType{
	"dxgi":   DXGI,
	"bitblt": BitBlt,
	"scrap":  Scrap,
}

type StripType int

const (
	RGBW StripType = ws281x.SK6812StripRGBW
	RBGW StripType = ws281x.SK6812StripRBGW
	GRBW StripType = ws281x.SK6812StripGRBW
	GBRW StripType = ws281x.SK6812StrioGBRW
	BRGW StripType = ws281x.SK6812StrioBRGW
	BGRW StripType = ws281x.SK6812StripBGRW
	RGB  StripType = ws281x.WS2811StripRGB
	RBG  StripType = ws281x.WS2811StripRBG
	GRB  StripType = ws281x.WS2811StripGRB
	GBR  StripType = ws281x.WS2811StripGBR
	BRG  StripType = ws281x.WS2811StripBRG
	BGR  StripType = ws281x.WS2811StripBGR
)

var stripTypes = map[string]StripType{
	"rgbw": RGBW,
	"rbgw": RBGW,
	"grbw": GRBW,
	"gbrw": GBRW,
	"brgw": BRGW,
	"bgrw": BGRW,
	"rgb":  RGB,
	"rbg":  RBG,
	"grb":  GRB,
	"gbr":  GBR,
	"brg":  BRG,
	"bgr":  BGR,
}

func (a *App) validateConfig(c config.Config) error {
	_, ok := controller.Modes[c.DefaultMode]
	if !ok {
		return fmt.Errorf("invalid default mode")
	}

	_, ok = capturerTypes[c.CapturerType]
	if !ok {
		return fmt.Errorf("invalid capturer type")
	}

	err := a.validateServer(c.Server)
	if err != nil {
		return err
	}

	err = a.validateDisplays(c.Displays)
	if err != nil {
		return err
	}

	return nil
}

func (a *App) validateServer(srv config.Server) error {
	ip := net.ParseIP(srv.Host)
	if ip == nil {
		return fmt.Errorf("invalid server IP")
	}

	if srv.Port < 1 || srv.Port > 65535 {
		return fmt.Errorf("invalid server port")
	}

	if srv.Leds < 1 || srv.Leds > 1024 {
		return fmt.Errorf("invalid server LED count")
	}

	_, ok := stripTypes[srv.StripType]
	if !ok {
		return fmt.Errorf("invalid server strip type")
	}

	if srv.GpioPin < 0 || srv.GpioPin > 27 {
		return fmt.Errorf("invalid server GPIO pin")
	}

	if srv.Brightness < 0 || srv.Brightness > 255 {
		return fmt.Errorf("invalid server brightness")
	}

	return nil
}

func (a *App) validateDisplays(displays []config.Display) error {
	for i, d := range displays {
		if d.Leds < 1 || d.Leds > 1024 {
			return fmt.Errorf("invalid LED count for display %d", i)
		}

		if d.Width < 1 || d.Width > 7680 {
			return fmt.Errorf("invalid width for display %d", i)
		}

		if d.Height < 1 || d.Height > 4320 {
			return fmt.Errorf("invalid width for display %d", i)
		}

		v1 := validateBounds(d.Width, d.Height, d.Bounds.From.X, d.Bounds.From.Y)
		if !v1 {
			return fmt.Errorf("invalid bounds for display %d (from)", i)
		}

		v2 := validateBounds(d.Width, d.Height, d.Bounds.To.X, d.Bounds.To.Y)
		if !v2 {
			return fmt.Errorf("invalid bounds for display %d (to)", i)
		}
	}

	return nil
}

func (a *App) applyConfig(c config.Config) (err error) {
	switch CapturerType(c.CapturerType) {
	case DXGI:
		a.displayRepository, err = dxgi.New()
		if err != nil {
			return err
		}
	case BitBlt:
		a.displayRepository = bitblt.New()
	case Scrap:
		a.displayRepository, err = scrap.New()
		if err != nil {
			return err
		}
	}

	a.DefaultMode = controller.Mode(c.DefaultMode)
	a.Host = c.Server.Host
	a.Port = c.Server.Port
	a.Leds = c.Server.Leds
	a.StripType = stripTypes[c.Server.StripType]
	a.GpioPin = c.Server.GpioPin
	a.Brightness = c.Server.Brightness

	var displays []Display
	for i, d := range c.Displays {
		fromOffset := calculateOffset(d.Width, d.Height, d.Bounds.From.X, d.Bounds.From.Y)
		toOffset := calculateOffset(d.Width, d.Height, d.Bounds.To.X, d.Bounds.To.Y)

		size := getPixSliceSize(d.Width, d.Height, fromOffset, toOffset)

		displays = append(
			displays,
			Display{
				Id:           i,
				Width:        d.Width,
				Height:       d.Height,
				Leds:         d.Leds,
				BoundsOffset: fromOffset,
				BoundsSize:   size,
			},
		)
	}

	return nil
}
