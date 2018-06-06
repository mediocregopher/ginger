package main

import (
	"github.com/mediocregopher/ginger/gim/geo"
	"github.com/mediocregopher/ginger/gim/terminal"
)

type line struct {
	from, to   *box
	fromI, toI int
	body       string
}

// given the "primary" direction the line should be headed, picks a possible
// secondary one which may be used to detour along the path in order to reach
// the destination (in the case that the two boxes are diagonal from each other)
func secondaryDir(flowDir, start, end geo.XY) geo.XY {
	var perpDir geo.XY
	perpDir[0], perpDir[1] = flowDir[1], flowDir[0]
	return end.Sub(start).Mul(perpDir.Abs()).Unit()
}

func (l line) draw(buf *terminal.Buffer, flowDir, secFlowDir geo.XY) {
	from, to := *(l.from), *(l.to)

	start := from.rect().Edge(flowDir, secFlowDir)[0].Add(secFlowDir.Scale(l.fromI*2 + 1))
	end := to.rect().Edge(flowDir.Inv(), secFlowDir)[0].Add(secFlowDir.Scale(l.toI*2 + 1))
	dirSec := secondaryDir(flowDir, start, end)
	mid := start.Midpoint(end, rounder)

	along := func(xy, dir geo.XY) int {
		if dir[0] != 0 {
			return xy[0]
		}
		return xy[1]
	}

	// collect the points along the line into an array
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

	// draw each point
	for i, pt := range pts {
		var r rune
		switch {
		case i == 0:
			r = terminal.SingleLine.Perpendicular(flowDir)
		case i == len(pts)-1:
			r = terminal.SingleLine.Arrow(flowDir)
		default:
			prev, next := pts[i-1], pts[i+1]
			r = terminal.SingleLine.Segment(prev.Sub(pt), next.Sub(pt))
		}
		buf.SetPos(pt)
		buf.WriteRune(r)
	}

	// draw the body
	if l.body != "" {
		bodyPos := mid.Add(geo.Left.Scale(len(l.body) / 2))
		buf.SetPos(bodyPos)
		buf.WriteString(l.body)
	}
}
