package main

import (
	"math/rand"
	"time"

	"github.com/mediocregopher/ginger/gg"
	"github.com/mediocregopher/ginger/gim/geo"
	"github.com/mediocregopher/ginger/gim/terminal"
	"github.com/mediocregopher/ginger/gim/view"
)

// TODO be able to draw circular graphs
// TODO audit all steps, make sure everything is deterministic
// TODO self-edges

//const (
//	framerate   = 10
//	frameperiod = time.Second / time.Duration(framerate)
//)

//func debugf(str string, args ...interface{}) {
//	if !strings.HasSuffix(str, "\n") {
//		str += "\n"
//	}
//	fmt.Fprintf(os.Stderr, str, args...)
//}

func mkGraph() (*gg.Graph, gg.Value) {
	a := gg.NewValue("a")
	aE0 := gg.NewValue("aE0")
	aE1 := gg.NewValue("aE1")
	aE2 := gg.NewValue("aE2")
	aE3 := gg.NewValue("aE3")
	b0 := gg.NewValue("b0")
	b1 := gg.NewValue("b1")
	b2 := gg.NewValue("b2")
	b3 := gg.NewValue("b3")
	oaE0 := gg.ValueOut(a, aE0)
	oaE1 := gg.ValueOut(a, aE1)
	oaE2 := gg.ValueOut(a, aE2)
	oaE3 := gg.ValueOut(a, aE3)
	g := gg.Null
	g = g.AddValueIn(oaE0, b0)
	g = g.AddValueIn(oaE1, b1)
	g = g.AddValueIn(oaE2, b2)
	g = g.AddValueIn(oaE3, b3)

	c := gg.NewValue("c")
	empty := gg.NewValue("")
	jE := gg.JunctionOut([]gg.OpenEdge{
		gg.ValueOut(b0, empty),
		gg.ValueOut(b1, empty),
		gg.ValueOut(b2, empty),
		gg.ValueOut(b3, empty),
	}, gg.NewValue("jE"))
	g = g.AddValueIn(jE, c)

	// TODO this really fucks it up
	//d := gg.NewValue("d")
	//deE := gg.ValueOut(d, gg.NewValue("deE"))
	//g = g.AddValueIn(deE, gg.NewValue("e"))

	return g, c
}

//func mkGraph() *gg.Graph {
//	g := gg.Null
//	g = g.AddValueIn(gg.ValueOut(str("a"), str("e")), str("b"))
//	return g
//}

func main() {
	rand.Seed(time.Now().UnixNano())
	term := terminal.New()
	wSize := term.WindowSize()
	center := geo.Zero.Midpoint(wSize)

	g, start := mkGraph()
	view := view.New(g, start, geo.Right, geo.Down)
	viewBuf := terminal.NewBuffer()
	view.Draw(viewBuf)

	buf := terminal.NewBuffer()
	buf.DrawBufferCentered(center, viewBuf)

	term.Clear()
	term.WriteBuffer(geo.Zero, buf)
	term.SetPos(wSize.Add(geo.XY{0, -1}))
	term.Draw()
}
