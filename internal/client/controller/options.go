package controller

import (
	"ledctl3/internal/client/interfaces"
)

type Option func(*Controller) error

func WithLedsCount(leds int) Option {
	return func(ctl *Controller) error {
		ctl.leds = leds
		return nil
	}
}

func WithDisplayVisualizer(visualizer interfaces.Visualizer) Option {
	return func(ctl *Controller) error {
		ctl.displayVisualizer = visualizer
		return nil
	}
}

func WithAudioVisualizer(visualizer interfaces.Visualizer) Option {
	return func(ctl *Controller) error {
		ctl.audioVisualizer = visualizer
		return nil
	}
}
