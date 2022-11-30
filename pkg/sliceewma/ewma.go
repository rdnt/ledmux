package sliceewma

import (
	"github.com/VividCortex/ewma"
)

type MovingAverage interface {
	Add([]float64)
	Value() []float64
	Set([]float64)
}

type SliceMovingAverage struct {
	values []ewma.MovingAverage
}

func NewMovingAverage(size int, age ...float64) MovingAverage {
	vals := make([]ewma.MovingAverage, size)

	for i := 0; i < size; i++ {
		vals[i] = ewma.NewMovingAverage(age...)
	}

	return &SliceMovingAverage{
		values: vals,
	}
}

func (m *SliceMovingAverage) Add(values []float64) {
	for i := 0; i < len(m.values); i++ {
		m.values[i].Add(values[i])
	}
}

func (m *SliceMovingAverage) Value() []float64 {
	values := make([]float64, len(m.values))

	for i := 0; i < len(m.values); i++ {
		values[i] = m.values[i].Value()
	}

	return values
}

func (m *SliceMovingAverage) Set(values []float64) {
	for i := 0; i < len(m.values); i++ {
		m.values[i].Set(values[i])
	}
}
