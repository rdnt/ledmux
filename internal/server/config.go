package application

import (
	"errors"
	"fmt"

	"ledctl3/internal/pkg/strip"
	"ledctl3/internal/server/config"
)

func validateConfig(c config.Config) error {
	_, err := strip.Parse(c.StripType)
	if err != nil {
		return err
	}

	err = validateCalibration(c.Calibration)
	if err != nil {
		return err
	}

	return nil
}

func validateCalibration(calib []config.Calibration) error {
	calibs := map[int]bool{}

	for _, c := range calib {
		if c.Start < 0 || c.End < 0 {
			return errors.New("calibration index out of range")
		}

		if c.Start > c.End {
			return errors.New("calibration start must be less or equal to calibration end")
		}

		if c.Red < 0 || c.Red > 1 {
			return errors.New("calibration factor out of range")
		}

		if c.Blue < 0 || c.Blue > 1 {
			return errors.New("calibration factor out of range")
		}

		if c.Green < 0 || c.Green > 1 {
			return errors.New("calibration factor out of range")
		}

		if c.White < 0 || c.White > 1 {
			return errors.New("calibration factor out of range")
		}

		for i := c.Start; i <= c.End; i++ {
			_, ok := calibs[i]
			if ok {
				return errors.New("duplicate calibration index")
			}

			calibs[i] = true
		}
	}

	return nil
}

func (a *Application) applyConfig(c config.Config) (err error) {
	offset := 0
	segs := []Segment{}

	for _, seg := range c.Segments {
		segs = append(segs, Segment{
			id:    seg.Id,
			leds:  seg.Leds,
			start: offset,
			end:   offset + seg.Leds,
		})

		offset += seg.Leds
	}

	a.leds = offset
	a.gpioPin = c.GpioPin
	a.stripType = c.StripType
	a.brightness = c.Brightness
	a.segments = segs

	a.calibration = map[int]Calibration{}

	fmt.Println(a.leds)
	for _, c := range c.Calibration {
		if c.Start > a.leds || c.End > a.leds {
			return errors.New("calibration index out of range")
		}

		for i := c.Start; i <= c.End; i++ {
			a.calibration[i] = Calibration{
				Red:   c.Red,
				Green: c.Green,
				Blue:  c.Blue,
				White: c.White,
			}
		}
	}

	return nil
}
