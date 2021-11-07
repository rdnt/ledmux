package controller

import (
	"ledctl3/internal/client/interfaces"
)

type Option func(*Controller) error

func WithLedsCount(leds int) Option {
	return func(c *Controller) error {
		c.leds = leds
		return nil
	}
}

func WithDisplayVisualizer(visualizer interfaces.Visualizer) Option {
	return func(c *Controller) error {
		c.displayVisualizer = visualizer
		return nil
	}
}

func WithAudioVisualizer(visualizer interfaces.Visualizer) Option {
	return func(c *Controller) error {
		c.audioVisualizer = visualizer
		return nil
	}
}
