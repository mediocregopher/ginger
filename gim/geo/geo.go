// Package geo implements basic geometric concepts used by gim
package geo

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

// Div returns the results of dividing the two XYs' field individually, using
// the Rounder to resolve floating results
func (xy XY) Div(xy2 XY, r Rounder) XY {
	xyf, xy2f := xy.toF64(), xy2.toF64()
	return XY{
		r.Round(xyf[0] / xy2f[0]),
		r.Round(xyf[1] / xy2f[1]),
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

func (xy XY) toF64() [2]float64 {
	return [2]float64{
		float64(xy[0]),
		float64(xy[1]),
	}
}

// Midpoint returns the midpoint between the two XYs. The rounder indicates what
// to do about non-whole values when they're come across
func (xy XY) Midpoint(xy2 XY, r Rounder) XY {
	return xy.Add(xy2.Sub(xy).Div(XY{2, 2}, r))
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
