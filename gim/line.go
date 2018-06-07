package main

import (
	"github.com/mediocregopher/ginger/gim/geo"
	"github.com/mediocregopher/ginger/gim/terminal"
)

type line struct {
	from, to   *box
	fromI, toI int
	bodyBuf    *terminal.Buffer
}

func (l line) draw(buf *terminal.Buffer, flowDir, secFlowDir geo.XY) {
	from, to := *(l.from), *(l.to)
	start := from.rect().Edge(flowDir, secFlowDir)[0].Add(secFlowDir.Scale(l.fromI*2 + 1))
	end := to.rect().Edge(flowDir.Inv(), secFlowDir)[0]
	end = end.Add(flowDir.Inv())
	end = end.Add(secFlowDir.Scale(l.toI*2 + 1))

	buf.SetPos(start)
	buf.WriteRune(terminal.SingleLine.Perpendicular(flowDir))
	buf.DrawLine(start.Add(flowDir), end.Add(flowDir.Inv()), flowDir, terminal.SingleLine)
	buf.SetPos(end)
	buf.WriteRune(terminal.SingleLine.Arrow(flowDir))

	// draw the body
	if l.bodyBuf != nil {
		mid := start.Midpoint(end)
		bodyBufRect := geo.Rect{Size: l.bodyBuf.Size()}
		buf.DrawBuffer(bodyBufRect.Centered(mid).TopLeft, l.bodyBuf)
	}
}
