package main

import (
	"fmt"

	"github.com/mediocregopher/ginger/gg"
	"github.com/mediocregopher/ginger/gim/geo"
	"github.com/mediocregopher/ginger/gim/terminal"
)

type box struct {
	topLeft       geo.XY
	flowDir       geo.XY
	numIn, numOut int
	bodyBuf       *terminal.Buffer

	transparent bool
}

func boxFromVertex(v *gg.Vertex, flowDir geo.XY) box {
	b := box{
		flowDir: flowDir,
		numIn:   len(v.In),
		numOut:  len(v.Out),
	}
	if v.VertexType == gg.ValueVertex {
		b.bodyBuf = terminal.NewBuffer()
		b.bodyBuf.WriteString(v.Value.V.(string))
	}
	return b
}

func (b box) rect() geo.Rect {
	var bodyRect geo.Rect
	if b.bodyBuf != nil {
		bodyRect.Size = b.bodyBuf.Size().Add(geo.XY{2, 2})
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

func (b box) draw(buf *terminal.Buffer) {
	rect := b.rect()
	buf.DrawRect(rect, terminal.SingleLine)

	if b.bodyBuf != nil {
		center := rect.Center(rounder)
		bodyBufRect := geo.Rect{Size: b.bodyBuf.Size()}
		buf.DrawBuffer(bodyBufRect.Centered(center, rounder).TopLeft, b.bodyBuf)
	}
}
