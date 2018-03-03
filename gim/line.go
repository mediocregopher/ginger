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

type line struct {
	from, to *box
	toI      int
}

// given the "primary" direction the line should be headed, picks a possible
// secondary one which may be used to detour along the path in order to reach
// the destination (in the case that the two boxes are diagonal from each other)
func (l line) secondaryDir(primary geo.XY) geo.XY {
	fromRect, toRect := l.from.rect(), l.to.rect()
	rels := make([]int, len(geo.Units))
	for i, dir := range geo.Units {
		rels[i] = toRect.EdgeCoord(dir.Inv()) - fromRect.EdgeCoord(dir)
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

//func (l line) startEnd(flowDir, secFlowDir geo.XY) (geo.XY, geo.XY) {
//	from, to := *(l.from), *(l.to)
//	start := from.rect().EdgeMidpoint(flowDir, rounder) // ezpz
//}

func (l line) draw(term *terminal.Terminal, flowDir, secFlowDir geo.XY) {
	from, to := *(l.from), *(l.to)
	dirSec := l.secondaryDir(flowDir)

	flowDirInv := flowDir.Inv()
	start := from.rect().Edge(flowDir, secFlowDir).Midpoint(rounder)

	endSlot := l.toI*2 + 1
	endSlotXY := geo.XY{endSlot, endSlot}
	end := to.rect().Edge(flowDirInv, secFlowDir)[0].Add(secFlowDir.Mul(endSlotXY))

	mid := start.Midpoint(end, rounder)

	along := func(xy, dir geo.XY) int {
		if dir[0] != 0 {
			return xy[0]
		}
		return xy[1]
	}

	var pts []geo.XY
	midPrim := along(mid, flowDir)
	endSec := along(end, dirSec)
	for curr := start; curr != end; {
		pts = append(pts, curr)
		if prim := along(curr, flowDir); prim == midPrim {
			if sec := along(curr, dirSec); sec != endSec {
				curr = curr.Add(dirSec)
				continue
			}
		}
		curr = curr.Add(flowDir)
	}

	for i, pt := range pts {
		var str string
		switch {
		case i == 0:
			str = edgeSegments[flowDir]
		case i == len(pts)-1:
			str = arrows[flowDir]
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
