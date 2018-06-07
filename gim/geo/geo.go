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

// Units is the set of unit vectors
var Units = []XY{
	Up,
	Down,
	Left,
	Right,
}

func (xy XY) toF64() [2]float64 {
	return [2]float64{
		float64(xy[0]),
		float64(xy[1]),
	}
}

func abs(i int) int {
	if i < 0 {
		return i * -1
	}
	return i
}

// Abs returns the XY with all fields made positive, if they weren't already
func (xy XY) Abs() XY {
	return XY{abs(xy[0]), abs(xy[1])}
}

// Unit returns the XY with each field divided by its absolute value (i.e.
// scaled down to 1 or -1). Fields which are 0 are left alone
func (xy XY) Unit() XY {
	for i := range xy {
		if xy[i] > 0 {
			xy[i] = 1
		} else if xy[i] < 0 {
			xy[i] = -1
		}
	}
	return xy
}

// Len returns the length (aka magnitude) of the XY as a vector.
func (xy XY) Len() int {
	if xy[0] == 0 {
		return abs(xy[1])
	} else if xy[1] == 0 {
		return abs(xy[0])
	}

	xyf := xy.toF64()
	lf := math.Sqrt((xyf[0] * xyf[0]) + (xyf[1] * xyf[1]))
	return Rounder.Round(lf)
}

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

// Div returns the results of dividing the two XYs' field individually.
func (xy XY) Div(xy2 XY) XY {
	xyf, xy2f := xy.toF64(), xy2.toF64()
	return XY{
		Rounder.Round(xyf[0] / xy2f[0]),
		Rounder.Round(xyf[1] / xy2f[1]),
	}
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

// Midpoint returns the midpoint between the two XYs.
func (xy XY) Midpoint(xy2 XY) XY {
	return xy.Add(xy2.Sub(xy).Div(XY{2, 2}))
}

// Min returns an XY whose fields are the minimum values of the two XYs'
// fields compared individually
func (xy XY) Min(xy2 XY) XY {
	for i := range xy {
		if xy2[i] < xy[i] {
			xy[i] = xy2[i]
		}
	}
	return xy
}

// Max returns an XY whose fields are the Maximum values of the two XYs'
// fields compared individually
func (xy XY) Max(xy2 XY) XY {
	for i := range xy {
		if xy2[i] > xy[i] {
			xy[i] = xy2[i]
		}
	}
	return xy
}
