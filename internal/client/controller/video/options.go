package video

type Option func(p *Visualizer) error

func WithLedsCount(leds int) Option {
	return func(p *Visualizer) error {
		p.leds = leds
		return nil
	}
}

func WithDisplayRepository(displays DisplayRepository) Option {
	return func(p *Visualizer) error {
		p.displayRepo = displays
		return nil
	}
}

func WithDisplayConfig(cfg [][]DisplayConfig) Option {
	return func(p *Visualizer) error {
		p.displayConfigs = cfg
		return nil
	}
}
