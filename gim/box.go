package main

import (
	"fmt"
	"strings"

	"github.com/mediocregopher/ginger/gg"
	"github.com/mediocregopher/ginger/gim/geo"
	"github.com/mediocregopher/ginger/gim/terminal"
)

const (
	boxBorderHoriz = iota
	boxBorderVert
	boxBorderTL
	boxBorderTR
	boxBorderBL
	boxBorderBR
)

var boxDefault = []string{
	"─",
	"│",
	"┌",
	"┐",
	"└",
	"┘",
}

type box struct {
	topLeft       geo.XY
	flowDir       geo.XY
	numIn, numOut int
	body          string

	transparent bool
}

func boxFromVertex(v *gg.Vertex, flowDir geo.XY) box {
	b := box{
		flowDir: flowDir,
		numIn:   len(v.In),
		numOut:  len(v.Out),
	}
	if v.VertexType == gg.Value {
		b.body = string(v.Value.(str))
	}
	return b
}

func (b box) bodyLines() []string {
	lines := strings.Split(b.body, "\n")
	// if the last line is empty don't include it, it means there was a trailing
	// newline (or the whole string is empty)
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func (b box) bodySize() geo.XY {
	var size geo.XY
	for _, line := range b.bodyLines() {
		size[1]++
		if l := len(line); l > size[0] {
			size[0] = l
		}
	}

	return size
}

func (b box) rect() geo.Rect {
	bodyRect := geo.Rect{
		Size: b.bodySize().Add(geo.XY{2, 2}),
	}

	var edgesRect geo.Rect
	{
		var neededByEdges int
		if b.numIn > b.numOut {
			neededByEdges = b.numIn*2 + 1
		} else {
			neededByEdges = b.numOut*2 + 1
		}

		switch b.flowDir {
		case geo.Left, geo.Right:
			edgesRect.Size = geo.XY{neededByEdges, 2}
		case geo.Up, geo.Down:
			edgesRect.Size = geo.XY{2, neededByEdges}
		default:
			panic(fmt.Sprintf("unknown flowDir: %#v", b.flowDir))
		}
	}

	return bodyRect.Union(edgesRect).Translate(b.topLeft)
}

func (b box) bodyRect() geo.Rect {
	center := b.rect().Center(rounder)
	return geo.Rect{Size: b.bodySize()}.Centered(center, rounder)
}

func (b box) draw(term *terminal.Terminal) {
	chars := boxDefault
	rect := b.rect()
	pos := rect.TopLeft
	w, h := rect.Size[0], rect.Size[1]

	// draw top line
	term.MoveCursorTo(pos)
	term.Printf(chars[boxBorderTL])
	for i := 0; i < w-2; i++ {
		term.Printf(chars[boxBorderHoriz])
	}
	term.Printf(chars[boxBorderTR])
	pos[1]++

	// draw vertical lines
	for i := 0; i < h-2; i++ {
		term.MoveCursorTo(pos)
		term.Printf(chars[boxBorderVert])
		if b.transparent {
			term.MoveCursorTo(pos.Add(geo.XY{w, 0}))
		} else {
			term.Printf(strings.Repeat(" ", w-2))
		}
		term.Printf(chars[boxBorderVert])
		pos[1]++
	}

	// draw bottom line
	term.MoveCursorTo(pos)
	term.Printf(chars[boxBorderBL])
	for i := 0; i < w-2; i++ {
		term.Printf(chars[boxBorderHoriz])
	}
	term.Printf(chars[boxBorderBR])

	// write out inner lines
	pos = b.bodyRect().TopLeft
	for _, line := range b.bodyLines() {
		term.MoveCursorTo(pos)
		term.Printf(line)
		pos[1]++
	}
}
