package client

import (
	"errors"
	"fmt"
	"image/color"

	"github.com/lucasb-eyer/go-colorful"
	"golang.org/x/exp/slices"

	"ledctl3/internal/client/config"
	"ledctl3/internal/client/controller"
	"ledctl3/internal/client/controller/video"
	"ledctl3/internal/client/controller/video/capturer/bitblt"
	"ledctl3/internal/client/controller/video/capturer/dxgi"
)

type CapturerType string

const (
	DXGI   CapturerType = "dxgi"
	BitBlt CapturerType = "bitblt"
)

var capturerTypes = map[string]CapturerType{
	"dxgi":   DXGI,
	"bitblt": BitBlt,
}

type StripType string

const (
	RGBW StripType = "rgbw"
	RBGW StripType = "rbgw"
	GRBW StripType = "grbw"
	GBRW StripType = "gbrw"
	BRGW StripType = "brgw"
	BGRW StripType = "bgrw"
	RGB  StripType = "rgb"
	RBG  StripType = "rbg"
	GRB  StripType = "grb"
	GBR  StripType = "gbr"
	BRG  StripType = "brg"
	BGR  StripType = "bgr"
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

func (a *Application) validateConfig(c config.Config) error {
	_, ok := controller.Modes[c.DefaultMode]
	if !ok {
		return fmt.Errorf("invalid default mode")
	}

	_, ok = capturerTypes[c.CaptureType]
	if !ok {
		return fmt.Errorf("invalid capturer type")
	}

	err := a.validateSegments(c.Segments)
	if err != nil {
		return err
	}

	err = a.validateServer(c.Server)
	if err != nil {
		return err
	}

	err = a.validateDisplayConfigs(c.Displays)
	if err != nil {
		return err
	}

	err = a.validateAudioConfig(c.Audio)
	if err != nil {
		return err
	}

	err = a.validateCalibration(c.Calibration)
	if err != nil {
		return err
	}

	return nil
}

func (a *Application) validateSegments(segs []config.Segment) error {
	for _, seg := range segs {
		if seg.Leds < 1 || seg.Leds > 1024 {
			return fmt.Errorf("invalid LED count for segment %d", seg.Id)
		}
	}

	return nil
}

func (a *Application) validateServer(srv config.Server) error {
	//ip := net.ParseIP(srv.Host)
	//if ip == nil {
	//	return fmt.Errorf("invalid server IP")
	//}

	if srv.Port < 1 || srv.Port > 65535 {
		return fmt.Errorf("invalid server port")
	}

	if srv.Leds < 1 || srv.Leds > 2048 {
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

func (a *Application) validateDisplayConfigs(displayConfigs [][]config.Display) error {
	for _, cfg := range displayConfigs {
		for i, d := range cfg {
			if d.Width < 1 || d.Width > 7680 {
				return fmt.Errorf("invalid width for display %d", i)
			}

			if d.Height < 1 || d.Height > 4320 {
				return fmt.Errorf("invalid width for display %d", i)
			}

			if d.Framerate <= 0 {
				return fmt.Errorf("invalid framerate for display %d", i)
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
	}

	return nil
}

func (a *Application) validateAudioConfig(cfg config.Audio) error {
	if len(cfg.Colors.Profiles) == 0 {
		return errors.New("a color profile is required")
	}

	if cfg.Colors.Selected == "" {
		return errors.New("the selected color profile is invalid")
	}

	var found bool

	names := []string{}
	for _, prof := range cfg.Colors.Profiles {
		if prof.Name == "" {
			return errors.New("color profile name is required")
		}

		if slices.Contains(names, prof.Name) {
			return errors.New("duplicate profile name")
		}
		names = append(names, prof.Name)

		if prof.Name == cfg.Colors.Selected {
			found = true
		}

		if len(prof.Colors) < 2 {
			return errors.New("a color profile requires a minimum of two colors")
		}

		for _, hex := range prof.Colors {
			_, err := colorful.Hex(hex)
			if err != nil {
				return err
			}
		}
	}

	if !found {
		return errors.New("the selected profile doesn't exist")
	}

	if cfg.WindowSize < 1 || cfg.WindowSize > 1000 {
		return errors.New("averaging window must be at least 1 and no more than 1000 frames")
	}

	if cfg.BlackPoint < 0 || cfg.BlackPoint >= 1 {
		return errors.New("black point has to be a floating point number in the range [0-1)")
	}

	return nil
}

func (a *Application) validateCalibration(calib []config.Calibration) error {
	calibs := map[int]bool{}

	for _, c := range calib {
		_, ok := calibs[c.Id]
		if ok {
			return errors.New("duplicate calibration found")
		}

		calibs[c.Id] = true
	}

	return nil
}

func (a *Application) applyConfig(c config.Config) (err error) {
	switch CapturerType(c.CaptureType) {
	case DXGI:
		a.Displays, err = dxgi.New()
		if err != nil {
			return err
		}
	case BitBlt:
		a.Displays = bitblt.New()
	}

	a.DefaultMode = controller.Mode(c.DefaultMode)
	a.Host = c.Server.Host
	a.Port = c.Server.Port
	a.Leds = c.Server.Leds
	a.StripType = stripTypes[c.Server.StripType]
	a.GpioPin = c.Server.GpioPin
	a.Brightness = c.Server.Brightness

	a.Segments = []Segment{}
	for _, s := range c.Segments {
		a.Segments = append(
			a.Segments, Segment{
				Id:   s.Id,
				Leds: s.Leds,
			},
		)
	}

	for i, cfg := range c.Displays {
		parsedCfg := []video.DisplayConfig{}

		for j, d := range cfg {
			fromOffset := calculateOffset(d.Width, d.Height, d.Bounds.From.X, d.Bounds.From.Y)
			toOffset := calculateOffset(d.Width, d.Height, d.Bounds.To.X, d.Bounds.To.Y)

			size := getPixSliceSize(d.Width, d.Height, fromOffset, toOffset)

			leds := 0
			for _, seg := range a.Segments {
				if seg.Id == d.Segment {
					leds = seg.Leds
				}
			}

			if leds == 0 {
				return fmt.Errorf("segment not found for display %d of config %d", j, i)
			}

			parsedCfg = append(
				parsedCfg, video.DisplayConfig{
					Id:           j,
					SegmentId:    d.Segment,
					Leds:         leds,
					Width:        d.Width,
					Height:       d.Height,
					Left:         d.Left,
					Top:          d.Top,
					Framerate:    d.Framerate,
					BoundsOffset: fromOffset,
					BoundsSize:   size,
					Bounds:       d.Bounds,
				},
			)
		}

		a.DisplayConfigs = append(a.DisplayConfigs, parsedCfg)
	}

	a.Colors = []color.Color{}

	for _, prof := range c.Audio.Colors.Profiles {
		if prof.Name == c.Audio.Colors.Selected {
			for _, hex := range prof.Colors {
				clr, err := colorful.Hex(hex)
				if err != nil {
					return err
				}

				a.Colors = append(a.Colors, clr)
			}

			break
		}
	}

	a.WindowSize = c.Audio.WindowSize
	a.BlackPoint = c.Audio.BlackPoint

	a.Calibration = map[int]Calibration{}
	
	for _, c := range c.Calibration {
		if c.Id > a.Leds {
			return errors.New("calibration index out of range")
		}

		a.Calibration[c.Id] = Calibration{
			Red:   c.Red,
			Green: c.Green,
			Blue:  c.Blue,
			White: c.White,
		}
	}

	return nil
}
