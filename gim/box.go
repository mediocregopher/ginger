package main

import (
	"fmt"
	"strings"

	"github.com/mediocregopher/ginger/gg"
	"github.com/mediocregopher/ginger/gim/geo"
	"github.com/mediocregopher/ginger/gim/terminal"
)

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
	if v.VertexType == gg.ValueVertex {
		b.body = v.Value.V.(string)
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

// TODO this is utterly broken, the terminal.Buffer should be used for this
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
			edgesRect.Size = geo.XY{2, neededByEdges}
		case geo.Up, geo.Down:
			edgesRect.Size = geo.XY{neededByEdges, 2}
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

func (b box) draw(buf *terminal.Buffer) {
	bodyBuf := terminal.NewBuffer()
	bodyBuf.WriteString(b.body)
	bodyBufRect := geo.Rect{Size: bodyBuf.Size()}

	rect := b.rect()
	buf.DrawRect(rect, terminal.SingleLine)

	center := rect.Center(rounder)
	buf.DrawBuffer(bodyBufRect.Centered(center, rounder).TopLeft, bodyBuf)
}
