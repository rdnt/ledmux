package ambilight

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

func WithDisplayRepository(displays interfaces.DisplayRepository) Option {
	return func(p *Visualizer) error {
		p.displays = displays
		return nil
	}
}

func WithDisplayConfig(cfg []DisplayConfig) Option {
	return func(p *Visualizer) error {
		p.displayCfg = cfg
		return nil
	}
}

