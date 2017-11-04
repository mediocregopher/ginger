// Package geo implements basic geometric concepts used by gim
package geo

import "math"

// XY describes a 2-dimensional position or vector. The origin of the
// 2-dimensional space is a 0,0, with the x-axis going to the left and the
// y-axis going down.
type XY [2]int

// Zero is the zero point, or a zero vector, depending on what you're doing
var Zero = XY{0, 0}

// Unit vectors
var (
	Up    = XY{0, -1}
	Down  = XY{0, 1}
	Left  = XY{-1, 0}
	Right = XY{1, 0}
)

// Add returns the result of adding the two XYs' fields individually
func (xy XY) Add(xy2 XY) XY {
	xy[0] += xy2[0]
	xy[1] += xy2[1]
	return xy
}

// Mul returns the result of multiplying the two XYs' fields individually
func (xy XY) Mul(xy2 XY) XY {
	xy[0] *= xy2[0]
	xy[1] *= xy2[1]
	return xy
}

// Scale returns the result of multiplying both of the XY's fields by the scalar
func (xy XY) Scale(scalar int) XY {
	return xy.Mul(XY{scalar, scalar})
}

// Inv inverses the XY, a shortcut for xy.Scale(-1)
func (xy XY) Inv() XY {
	return xy.Scale(-1)
}

// Sub subtracts xy2 from xy and returns the result. A shortcut for
// xy.Add(xy2.Inv())
func (xy XY) Sub(xy2 XY) XY {
	return xy.Add(xy2.Inv())
}

func round(f float64, r int) int {
	switch {
	case r < 0:
		f = math.Floor(f)
	case r == 0:
		if f < 0 {
			f = math.Ceil(f - 0.5)
		}
		f = math.Floor(f + 0.5)
	case r > 0:
		f = math.Ceil(f)
	}
	return int(f)
}

func (xy XY) toF64() [2]float64 {
	return [2]float64{
		float64(xy[0]),
		float64(xy[1]),
	}
}

// Midpoint returns the midpoint between the two XYs. The rounder indicates what
// to do about non-whole values when they're come across:
// - rounder < 0 : floor
// - rounder = 0 : round
// - rounder > 0 : ceil
func (xy XY) Midpoint(xy2 XY, rounder int) XY {
	xyf, xy2f := xy.toF64(), xy2.toF64()
	xf := xyf[0] + ((xy2f[0] - xyf[0]) / 2)
	yf := xyf[1] + ((xy2f[1] - xyf[1]) / 2)
	return XY{
		round(xf, rounder),
		round(yf, rounder),
	}
}
