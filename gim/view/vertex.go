package view

import (
	"github.com/mediocregopher/ginger/gim/geo"
	"github.com/mediocregopher/ginger/gim/terminal"
)

type edge struct {
	from, to   *vertex
	tail, head rune // if empty do directional segment char
	body       string
	switchback bool

	lineStyle terminal.LineStyle
}

type vertex struct {
	coord, pos geo.XY
	in, out    [][]*edge // top level is port index
	body       string

	// means it won't be drawn, and will be removed and have its in/out edges
	// spliced together into a single edge.
	ephemeral bool

	lineStyle terminal.LineStyle // if zero value don't draw border
}
