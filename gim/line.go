package main

import (
	"github.com/mediocregopher/ginger/gim/geo"
	"github.com/mediocregopher/ginger/gim/terminal"
)

var lineSegments = func() map[[2]geo.XY]string {
	m := map[[2]geo.XY]string{
		{geo.Left, geo.Right}: "─",
		{geo.Down, geo.Up}:    "│",
		{geo.Right, geo.Down}: "┌",
		{geo.Left, geo.Down}:  "┐",
		{geo.Right, geo.Up}:   "└",
		{geo.Left, geo.Up}:    "┘",
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

type line [2]*box

// given the "primary" direction the line should be headed, picks a possible
// secondary one which may be used to detour along the path in order to reach
// the destination (in the case that the two boxes are diagonal from each other)
func (l line) secondaryDir(primary geo.XY) geo.XY {
	fromRect, toRect := l[0].rect(), l[1].rect()
	rels := make([]int, len(geo.Units))
	for i, dir := range geo.Units {
		rels[i] = toRect.Edge(dir.Inv()) - fromRect.Edge(dir)
		if dir == geo.Up || dir == geo.Left {
			rels[i] *= -1
		}
	}

	var secondary geo.XY
	var secondaryMax int
	var secondarySet bool
	for i, rel := range rels {
		if geo.Units[i] == primary {
			continue
		} else if geo.Units[i][0] == 0 && primary[0] == 0 {
			continue
		} else if geo.Units[i][1] == 0 && primary[1] == 0 {
			continue
		} else if !secondarySet || rel > secondaryMax {
			secondary = geo.Units[i]
			secondaryMax = rel
			secondarySet = true
		}
	}

	return secondary
}

func (l line) draw(term *terminal.Terminal, dir geo.XY) {
	from, to := *l[0], *l[1]
	dirSec := l.secondaryDir(dir)

	dirInv := dir.Inv()
	start := from.rect().EdgeMidpoint(dir, rounder)
	end := to.rect().EdgeMidpoint(dirInv, rounder)
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
