package main

import (
	"fmt"

	"github.com/mediocregopher/ginger/gim/geo"
	"github.com/mediocregopher/ginger/gim/terminal"
)

// boxEdgeAdj returns the midpoint of a box's edge, using the given direction
// (single-dimension unit-vector) to know which edge to look at.
func boxEdgeAdj(box box, dir geo.XY) geo.XY {
	boxRect := box.rect()
	var a, b geo.XY
	switch dir {
	case geo.Up:
		a, b = boxRect.Corner(geo.Left, geo.Up), boxRect.Corner(geo.Right, geo.Up)
	case geo.Down:
		a, b = boxRect.Corner(geo.Left, geo.Down), boxRect.Corner(geo.Right, geo.Down)
	case geo.Left:
		a, b = boxRect.Corner(geo.Left, geo.Up), boxRect.Corner(geo.Left, geo.Down)
	case geo.Right:
		a, b = boxRect.Corner(geo.Right, geo.Up), boxRect.Corner(geo.Right, geo.Down)
	default:
		panic(fmt.Sprintf("unsupported direction: %#v", dir))
	}

	mid := a.Midpoint(b, rounder)
	return mid
}

var dirs = []geo.XY{
	geo.Up,
	geo.Down,
	geo.Left,
	geo.Right,
}

// boxesRelDir returns the "best" direction between from and to. Returns
// geo.Zero if they overlap. It also returns the secondary direction. E.g. Down
// and Left. The secondary direction will never be zero if primary is given,
// even if the two boxes are in-line
func boxesRelDir(from, to box) (geo.XY, geo.XY) {
	fromRect, toRect := from.rect(), to.rect()
	rels := make([]int, len(dirs))
	for i, dir := range dirs {
		rels[i] = toRect.Edge(dir.Inv()) - fromRect.Edge(dir)
		if dir == geo.Up || dir == geo.Left {
			rels[i] *= -1
		}
	}

	// find primary
	var primary geo.XY
	var primaryMax int
	for i, rel := range rels {
		if rel < 0 {
			continue
		} else if rel > primaryMax || i == 0 {
			primary = dirs[i]
			primaryMax = rel
		}
	}

	// if all rels were negative the boxes are overlapping, return zeros
	if primary == geo.Zero {
		return geo.Zero, geo.Zero
	}

	// now find secondary, which must be perpendicular to primary
	var secondary geo.XY
	var secondaryMax int
	var secondarySet bool
	for i, rel := range rels {
		if dirs[i] == primary {
			continue
		} else if dirs[i][0] == 0 && primary[0] == 0 {
			continue
		} else if dirs[i][1] == 0 && primary[1] == 0 {
			continue
		} else if !secondarySet || rel > secondaryMax {
			secondary = dirs[i]
			secondaryMax = rel
			secondarySet = true
		}
	}

	return primary, secondary
}

var lineSegments = func() map[[2]geo.XY]string {
	m := map[[2]geo.XY]string{
		{{-1, 0}, {1, 0}}:  "─",
		{{0, 1}, {0, -1}}:  "│",
		{{1, 0}, {0, 1}}:   "┌",
		{{-1, 0}, {0, 1}}:  "┐",
		{{1, 0}, {0, -1}}:  "└",
		{{-1, 0}, {0, -1}}: "┘",
	}

	// the inverse segments use the same characters
	for seg, str := range m {
		seg[0], seg[1] = seg[1], seg[0]
		m[seg] = str
	}
	return m
}()

var edgeSegments = map[geo.XY]string{
	geo.Up:    "┴",
	geo.Down:  "┬",
	geo.Left:  "┤",
	geo.Right: "├",
}

// actual unicode arrows were fucking up my terminal, and they didn't even
// connect properly with the line segments anyway
var arrows = map[geo.XY]string{
	geo.Up:    "^",
	geo.Down:  "v",
	geo.Left:  "<",
	geo.Right: ">",
}

func basicLine(term *terminal.Terminal, from, to box) {
	dir, dirSec := boxesRelDir(from, to)

	// if the boxes overlap then don't draw anything
	if dir == geo.Zero {
		return
	}

	dirInv := dir.Inv()
	start := boxEdgeAdj(from, dir)
	end := boxEdgeAdj(to, dirInv)
	mid := start.Midpoint(end, rounder)

	along := func(xy, dir geo.XY) int {
		if dir[0] != 0 {
			return xy[0]
		}
		return xy[1]
	}

	var pts []geo.XY
	midPrim := along(mid, dir)
	endSec := along(end, dirSec)
	for curr := start; curr != end; {
		pts = append(pts, curr)
		if prim := along(curr, dir); prim == midPrim {
			if sec := along(curr, dirSec); sec != endSec {
				curr = curr.Add(dirSec)
				continue
			}
		}
		curr = curr.Add(dir)
	}

	for i, pt := range pts {
		var str string
		switch {
		case i == 0:
			str = edgeSegments[dir]
		case i == len(pts)-1:
			str = arrows[dir]
		default:
			prev, next := pts[i-1], pts[i+1]
			seg := [2]geo.XY{
				prev.Sub(pt),
				next.Sub(pt),
			}
			str = lineSegments[seg]
		}
		term.MoveCursorTo(pt)
		term.Printf(str)
	}
}
