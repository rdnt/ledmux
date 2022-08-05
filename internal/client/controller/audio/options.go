package audio

type Option func(p *Visualizer) error

type Options struct {
	Leds     int
	Segments []Segment
}

func WithLedsCount(leds int) Option {
	return func(p *Visualizer) error {
		p.leds = leds
		return nil
	}
}

func WithSegments(segs []Segment) Option {
	return func(p *Visualizer) error {
		p.maxLedCount = 0

		for _, seg := range segs {
			if seg.Leds > p.maxLedCount {
				p.maxLedCount = seg.Leds
			}
		}

		p.segments = segs
		return nil
	}
}
