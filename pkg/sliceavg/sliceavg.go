package sliceavg

type Average interface {
	Add(values []float64)
	Current() []float64
	Update(values []float64) []float64
}

type exponentialMovingAverage struct {
	prev             []float64
	oneMinusConstant float64
	constant         float64
}

func New(size int, start []float64, smoothing float64) Average {
	constant := smoothing / (1 + float64(size))

	return &exponentialMovingAverage{
		constant:         constant,
		oneMinusConstant: 1 - constant,
		prev:             start,
	}
}

func (avg *exponentialMovingAverage) Add(values []float64) {
	_ = avg.Update(values)
}

func (avg *exponentialMovingAverage) Current() []float64 {
	return avg.prev
}

func (avg *exponentialMovingAverage) Update(values []float64) []float64 {
	if len(values) != len(avg.prev) {
		panic("values length differ")
	}

	for i, next := range values {
		avg.prev[i] = avg.calculate(avg.prev[i], next)
	}

	return avg.prev
}

func (avg *exponentialMovingAverage) calculate(prev, next float64) float64 {
	return next*avg.constant + prev*avg.oneMinusConstant
}
