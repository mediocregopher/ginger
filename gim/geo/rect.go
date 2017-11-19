package geo

import (
	"fmt"
)

// Rect describes a rectangle based on the position of its top-left corner and
// size
type Rect struct {
	TopLeft XY
	Size    XY
}

// Edge returns the coordinate of the edge indicated by the given direction (Up,
// Down, Left, or Right). The coordinate will be for the axis applicable to the
// direction, so for Left/Right it will be the x coordinate and for Up/Down the
// y.
func (r Rect) Edge(dir XY) int {
	switch dir {
	case Up:
		return r.TopLeft[1]
	case Down:
		return r.TopLeft[1] + r.Size[1] - 1
	case Left:
		return r.TopLeft[0]
	case Right:
		return r.TopLeft[0] + r.Size[0] - 1
	default:
		panic(fmt.Sprintf("unsupported direction: %#v", dir))
	}
}

// Corner returns the position of the corner identified by the given directions
// (Left/Right, Up/Down)
func (r Rect) Corner(xDir, yDir XY) XY {
	switch {
	case r.Size[0] == 0 || r.Size[1] == 0:
		panic(fmt.Sprintf("rectangle with non-multidimensional size has no corners: %v", r.Size))
	case xDir == Left && yDir == Up:
		return r.TopLeft
	case xDir == Right && yDir == Up:
		return r.TopLeft.Add(r.Size.Mul(Right)).Add(XY{-1, 0})
	case xDir == Left && yDir == Down:
		return r.TopLeft.Add(r.Size.Mul(Down)).Add(XY{0, -1})
	case xDir == Right && yDir == Down:
		return r.TopLeft.Add(r.Size).Add(XY{-1, -1})
	default:
		panic(fmt.Sprintf("unsupported Corner args: %v, %v", xDir, yDir))
	}
}

func (r Rect) halfSize(rounder Rounder) XY {
	return r.Size.Div(XY{2, 2}, rounder)
}

// Center returns the centerpoint of the rectangle, using the given Rounder to
// resolve non-integers
func (r Rect) Center(rounder Rounder) XY {
	return r.TopLeft.Add(r.halfSize(rounder))
}

// Translate returns an instance of Rect which is the same as this one but
// translated by the given amount
func (r Rect) Translate(by XY) Rect {
	r.TopLeft = r.TopLeft.Add(by)
	return r
}

// Centered returns an instance of Rect which is this one but translated to be
// centered on the given point. It will use the given Rounder to resolve
// non-integers
func (r Rect) Centered(on XY, rounder Rounder) Rect {
	r.TopLeft = on.Sub(r.halfSize(rounder))
	return r
}

// Union returns the smallest Rect which encompasses the given Rect and the one
// being called upon.
func (r Rect) Union(r2 Rect) Rect {
	if r.Size == Zero {
		return r2
	} else if r2.Size == Zero {
		return r
	}

	tl := r.TopLeft.Min(r2.TopLeft)
	br := r.Corner(Right, Down).Max(r2.Corner(Right, Down))
	return Rect{
		TopLeft: tl,
		Size:    br.Sub(tl).Add(XY{1, 1}),
	}
}
