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

func (l line) draw(term *terminal.Terminal, flowDir, secFlowDir geo.XY) {
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

	// draw the body
	if l.body != "" {
		bodyPos := mid.Add(geo.Left.Scale(len(l.body) / 2))
		term.MoveCursorTo(bodyPos)
		term.Printf(l.body)
	}
}
