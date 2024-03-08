package vector

import (
	"image/color"
	"math"
	"sync/atomic"
)

const (
	X = iota
	Y
	Z
)

type Perspective struct {
	X float64
	Y float64
}

var perspective atomic.Pointer[Perspective]

func init() {
	perspective.Store(new(Perspective))
	var (
		minX, minY = -0.4, 0.3
		px, py     = minX, minY
	)
	perspective.Store(&Perspective{X: px, Y: py})
}

type V3D [3]float64

func (v V3D) Scale(factor float64) (scaled V3D) {
	for i := range v {
		scaled[i] = v[i] * factor
	}
	return
}

func (v V3D) Sub(other V3D) (sub V3D) {
	for i := range v {
		sub[i] = v[i] - other[i]
	}
	return
}

func (v V3D) Add(other V3D) (sum V3D) {
	for i := range v {
		sum[i] = v[i] + other[i]
	}
	return sum
}

func (v V3D) Middle(other V3D) (middle V3D) {
	for coord := range v {
		middle[coord] = (v[coord] + other[coord]) / 2.
	}
	return middle
}

func (v V3D) Distance(other V3D) (distance float64) {
	return math.Sqrt(
		math.Pow(v[X]-other[X], 2.) +
			math.Pow(v[Y]-other[Y], 2.) +
			math.Pow(v[Z]-other[Z], 2.))
}

func (v V3D) To2D() V2D {
	xP, yP := v[X], v[Y]
	xP += v[Z] * perspective.Load().X
	yP += v[Z] * perspective.Load().Y
	return V2D{xP, yP}
}

type Vector2DColor struct {
	Vec   V2D
	Color color.Color
}

func (v V3D) To2DColor() Vector2DColor {
	const (
		maxDistanceZ = float64(1500)
		scale        = maxDistanceZ / float64(255)
	)
	var red, green, blue uint8
	if v.Z() < 0 {
		red = uint8(v.Z() / scale)
	}
	if v.Z() > 0 {
		blue = uint8(v.Z() / scale)
	}
	green = uint8((maxDistanceZ - math.Abs(v.Z())) / scale)
	return Vector2DColor{
		Vec: v.To2D(),
		Color: color.RGBA{
			R: red,
			G: green,
			B: blue,
			A: 255,
		},
	}
}

func (v V3D) X() float64 {
	return v[X]
}

func (v V3D) Y() float64 {
	return v[Y]
}

func (v V3D) Z() float64 {
	return v[Z]
}

func (v V3D) RotateZ(degrees float64) (rotated V3D) {
	radians := degreesToRadians(degrees)
	cosTheta := math.Cos(radians) // x projection
	sinTheta := math.Sin(radians) // y projection

	return V3D{
		v[X]*cosTheta - v[Y]*sinTheta,
		v[X]*sinTheta + v[Y]*cosTheta,
		v[Z],
	}
}

func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180.0
}
