package main

import (
	"log"
	"time"

	"github.com/mediocregopher/ginger/gim/geo"
	"github.com/mediocregopher/ginger/gim/terminal"
)

func main() {
	b := terminal.NewBuffer()
	b.WriteString("this is fun")

	b.SetFGColor(terminal.Blue)
	b.SetBGColor(terminal.Green)
	b.SetPos(geo.XY{18, 0})
	b.WriteString("blue and green")

	b.ResetStyle()
	b.SetFGColor(terminal.Red)
	b.SetPos(geo.XY{3, 3})
	b.WriteString("red!!!")

	b.ResetStyle()
	b.SetFGColor(terminal.Blue)
	b.SetPos(geo.XY{20, 0})
	b.WriteString("boo")

	bcp := b.Copy()
	b.DrawBuffer(geo.XY{2, 2}, bcp)
	b.DrawBuffer(geo.XY{-1, 1}, bcp)

	brect := terminal.NewBuffer()
	brect.DrawRect(geo.Rect{Size: b.Size().Add(geo.XY{2, 2})}, terminal.SingleLine)
	log.Printf("b.Size:%v", b.Size())
	brect.DrawBuffer(geo.XY{1, 1}, b)

	t := terminal.New()
	p := geo.XY{0, 0}
	dirH, dirV := geo.Right, geo.Down
	wsize := t.WindowSize()
	for range time.Tick(time.Second / 15) {
		t.Clear()
		t.WriteBuffer(p, brect)
		t.Draw()

		brectSize := brect.Size()
		p = p.Add(dirH).Add(dirV)
		if p[0] < 0 || p[0]+brectSize[0] > wsize[0] {
			dirH = dirH.Scale(-1)
			p = p.Add(dirH.Scale(2))
		}
		if p[1] < 0 || p[1]+brectSize[1] > wsize[1] {
			dirV = dirV.Scale(-1)
			p = p.Add(dirV.Scale(2))
		}
	}
}
