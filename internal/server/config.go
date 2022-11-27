package application

import (
	"ledctl3/internal/pkg/strip"
	"ledctl3/internal/server/config"
)

func validateConfig(c config.Config) error {
	_, err := strip.Parse(c.StripType)
	if err != nil {
		return err
	}

	return nil
}

func (a *Application) applyConfig(c config.Config) (err error) {
	leds := 0
	segs := []Segment{}

	for _, seg := range c.Segments {
		segs = append(segs, Segment{
			id:    seg.Id,
			start: leds,
			end:   leds + seg.Leds,
		})

		leds += seg.Leds
	}

	a.leds = leds
	a.gpioPin = c.GpioPin
	a.stripType = c.StripType
	a.brightness = c.Brightness
	a.segments = segs

	return nil
}
