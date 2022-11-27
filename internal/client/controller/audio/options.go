package audio

import (
	"errors"
	"image/color"
)

type Option func(v *Visualizer) error

type Options struct {
	Leds     int
	Segments []Segment
}

func WithLedsCount(leds int) Option {
	return func(v *Visualizer) error {
		v.leds = leds
		return nil
	}
}

func WithSegments(segs []Segment) Option {
	return func(v *Visualizer) error {
		v.maxLedCount = 0

		for _, seg := range segs {
			if seg.Leds > v.maxLedCount {
				v.maxLedCount = seg.Leds
			}
		}

		v.segments = segs
		return nil
	}
}

// WithColors accepts an array of hex colors in the form #RRGGBB
func WithColors(colors ...color.Color) Option {
	return func(v *Visualizer) error {
		if len(colors) < 2 {
			return errors.New("minimum two colors required")
		}

		v.colors = colors
		return nil
	}
}

func WithWindowSize(size int) Option {
	return func(v *Visualizer) error {
		v.windowSize = size
		return nil
	}
}
