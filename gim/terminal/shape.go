package terminal

import (
	"fmt"
	"strings"

	"github.com/mediocregopher/ginger/gim/geo"
)

// SingleLine is a set of single-pixel-width lines.
var SingleLine = LineStyle{
	Horiz:       '─',
	Vert:        '│',
	TopLeft:     '┌',
	TopRight:    '┐',
	BottomLeft:  '└',
	BottomRight: '┘',
	PerpUp:      '┴',
	PerpDown:    '┬',
	PerpLeft:    '┤',
	PerpRight:   '├',
	ArrowUp:     '^',
	ArrowDown:   'v',
	ArrowLeft:   '<',
	ArrowRight:  '>',
}

// LineStyle defines a set of characters to use together when drawing lines and
// corners.
type LineStyle struct {
	Horiz, Vert rune

	// Corner characters, identified as corners of a rectangle
	TopLeft, TopRight, BottomLeft, BottomRight rune

	// Characters for a straight segment a perpendicular attached
	PerpUp, PerpDown, PerpLeft, PerpRight rune

	// Characters for pointing arrows
	ArrowUp, ArrowDown, ArrowLeft, ArrowRight rune
}

// Segment takes two different directions (i.e. geo.Up/Down/Left/Right) and
// returns the line character which points in both of those directions.
//
// For example, SingleLine.Segment(geo.Up, geo.Left) returns '┘'.
func (ls LineStyle) Segment(a, b geo.XY) rune {
	inner := func(a, b geo.XY) rune {
		type c struct{ a, b geo.XY }
		switch (c{a, b}) {
		case c{geo.Up, geo.Down}:
			return ls.Vert
		case c{geo.Left, geo.Right}:
			return ls.Horiz
		case c{geo.Down, geo.Right}:
			return ls.TopLeft
		case c{geo.Down, geo.Left}:
			return ls.TopRight
		case c{geo.Up, geo.Right}:
			return ls.BottomLeft
		case c{geo.Up, geo.Left}:
			return ls.BottomRight
		default:
			return 0
		}
	}
	if r := inner(a, b); r != 0 {
		return r
	} else if r = inner(b, a); r != 0 {
		return r
	}
	panic(fmt.Sprintf("invalid LineStyle.Segment directions: %v, %v", a, b))
}

// Perpendicular returns the line character for a perpendicular segment
// traveling in the given direction.
func (ls LineStyle) Perpendicular(dir geo.XY) rune {
	switch dir {
	case geo.Up:
		return ls.PerpUp
	case geo.Down:
		return ls.PerpDown
	case geo.Left:
		return ls.PerpLeft
	case geo.Right:
		return ls.PerpRight
	default:
		panic(fmt.Sprintf("invalid LineStyle.Perpendicular direction: %v", dir))
	}
}

// Arrow returns the arrow character for an arrow pointing in the given
// direction.
func (ls LineStyle) Arrow(dir geo.XY) rune {
	switch dir {
	case geo.Up:
		return ls.ArrowUp
	case geo.Down:
		return ls.ArrowDown
	case geo.Left:
		return ls.ArrowLeft
	case geo.Right:
		return ls.ArrowRight
	default:
		panic(fmt.Sprintf("invalid LineStyle.Arrow direction: %v", dir))
	}
}

// DrawRect draws the given Rect to the Buffer with the given LineStyle. The
// Rect's TopLeft field is used for its position.
//
// If Rect's Size is not at least 2x2 this does nothing.
func (b *Buffer) DrawRect(r geo.Rect, ls LineStyle) {
	if r.Size[0] < 2 || r.Size[1] < 2 {
		return
	}
	horiz := strings.Repeat(string(ls.Horiz), r.Size[0]-2)

	b.SetPos(r.TopLeft)
	b.WriteRune(ls.TopLeft)
	b.WriteString(horiz)
	b.WriteRune(ls.TopRight)

	for i := 0; i < r.Size[1]-2; i++ {
		b.SetPos(r.TopLeft.Add(geo.XY{0, i + 1}))
		b.WriteRune(ls.Vert)
		b.SetPos(r.TopLeft.Add(geo.XY{r.Size[0] - 1, i + 1}))
		b.WriteRune(ls.Vert)
	}

	b.SetPos(r.TopLeft.Add(geo.XY{0, r.Size[1] - 1}))
	b.WriteRune(ls.BottomLeft)
	b.WriteString(horiz)
	b.WriteRune(ls.BottomRight)
}

// DrawLine draws a line from the start point to the ending one, primarily
// moving in the given direction, using the given LineStyle to do so.
func (b *Buffer) DrawLine(start, end, dir geo.XY, ls LineStyle) {
	// given the "primary" direction the line should be headed, pick a possible
	// secondary one which may be used to detour along the path in order to
	// reach the destination (in the case that the two boxes are diagonal from
	// each other)
	var perpDir geo.XY
	perpDir[0], perpDir[1] = dir[1], dir[0]
	dirSec := end.Sub(start).Mul(perpDir.Abs()).Unit()
	mid := start.Midpoint(end)

	along := func(xy, dir geo.XY) int {
		if dir[0] != 0 {
			return xy[0]
		}
		return xy[1]
	}

	// collect the points along the line into an array
	var pts []geo.XY
	var curr geo.XY
	midPrim := along(mid, dir)
	endSec := along(end, dirSec)
	for curr = start; curr != end; {
		pts = append(pts, curr)
		if prim := along(curr, dir); prim == midPrim {
			if sec := along(curr, dirSec); sec != endSec {
				curr = curr.Add(dirSec)
				continue
			}
		}
		curr = curr.Add(dir)
	}
	pts = append(pts, curr) // appending end

	// draw each point
	for i, pt := range pts {
		var prev, next geo.XY
		switch {
		case i == 0:
			prev = pt.Add(dir.Inv())
			next = pts[i+1]
		case i == len(pts)-1:
			prev = pts[i-1]
			next = pt.Add(dir)
		default:
			prev, next = pts[i-1], pts[i+1]
		}
		b.SetPos(pt)
		b.WriteRune(ls.Segment(prev.Sub(pt), next.Sub(pt)))
	}
}
