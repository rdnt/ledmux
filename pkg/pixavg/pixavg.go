package pixavg

import "image/color"

type Average interface {
	Add(values []color.Color)
	Current() []color.Color
}

type exponentialMovingAverage struct {
	prev             []color.Color
	oneMinusConstant float64
	constant         float64
}

func New(size int, start []color.Color, smoothing float64) Average {
	constant := smoothing / (1 + float64(size))

	return &exponentialMovingAverage{
		constant:         constant,
		oneMinusConstant: 1 - constant,
		prev:             start,
	}
}

func (avg *exponentialMovingAverage) Add(values []color.Color) {
	if len(values) != len(avg.prev) {
		panic("values length differ")
	}

	for i, next := range values {
		r1, g1, b1, a1 := avg.prev[i].RGBA()
		r2, g2, b2, a2 := next.RGBA()

		r2 = avg.calculate(r1, r2)
		g2 = avg.calculate(g1, g2)
		b2 = avg.calculate(b1, b2)
		a2 = avg.calculate(a1, a2)

		avg.prev[i] = color.RGBA64{
			R: uint16(r2),
			G: uint16(g2),
			B: uint16(b2),
			A: uint16(a2),
		}
	}
}

func (avg *exponentialMovingAverage) Current() []color.Color {
	return avg.prev
}

func (avg *exponentialMovingAverage) calculate(prev, next uint32) uint32 {
	return uint32(float64(next)*avg.constant + float64(prev)*avg.oneMinusConstant)
}
