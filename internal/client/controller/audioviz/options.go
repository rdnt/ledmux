package audioviz

import (
	"ledctl3/internal/client/interfaces"
)

type Option func(p *Visualizer) error

func WithLedsCount(leds int) Option {
	return func(p *Visualizer) error {
		p.leds = leds
		return nil
	}
}

func WithAudioSource(source interfaces.AudioSource) Option {
	return func(p *Visualizer) error {
		p.source = source
		return nil
	}
}
