package vector

import (
	"math"
)

type V2D [2]float64

func (v V2D) Distance(other V2D) (distance float64) {
	return math.Sqrt(
		math.Pow(v[X]-other[X], 2.) +
			math.Pow(v[Y]-other[Y], 2.))
}

func (v V2D) Scale(factor float64) (scaled V2D) {
	for i := range v {
		scaled[i] = v[i] * factor
	}
	return
}

func (v V2D) X() float64 {
	return v[X]
}

func (v V2D) Y() float64 {
	return v[Y]
}
