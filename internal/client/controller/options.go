package controller

import (
	"ledctl3/internal/client/visualizer"
)

type Option func(*Controller) error

func WithLedsCount(leds int) Option {
	return func(ctl *Controller) error {
		ctl.leds = leds
		return nil
	}
}

func WithDisplayVisualizer(visualizer visualizer.Visualizer) Option {
	return func(ctl *Controller) error {
		ctl.displayVisualizer = visualizer
		return nil
	}
}

func WithAudioVisualizer(visualizer visualizer.Visualizer) Option {
	return func(ctl *Controller) error {
		ctl.audioVisualizer = visualizer
		return nil
	}
}

func WithSegmentsCount(count int) Option {
	return func(ctl *Controller) error {
		ctl.segmentCount = count
		return nil
	}
}
