package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/mediocregopher/ginger/gg"
	"github.com/mediocregopher/ginger/gim/geo"
	"github.com/mediocregopher/ginger/gim/terminal"
)

// Leave room for:
// - Changing the "flow" direction
// - Absolute positioning of some/all vertices

// TODO
// - edge values
// - be able to draw circular graphs
// - audit all steps, make sure everything is deterministic

const (
	framerate   = 10
	frameperiod = time.Second / time.Duration(framerate)
	rounder     = geo.Ceil
)

func debugf(str string, args ...interface{}) {
	if !strings.HasSuffix(str, "\n") {
		str += "\n"
	}
	fmt.Fprintf(os.Stderr, str, args...)
}

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
	//term.Reset()
	//term.HideCursor()

	g, start := mkGraph()
	v := view{
		g:           g,
		primFlowDir: geo.Right,
		secFlowDir:  geo.Down,
		start:       start,
		center:      geo.Zero.Midpoint(term.WindowSize(), rounder),
	}

	//for range time.Tick(frameperiod) {
	term.Reset()
	v.draw(term)
	term.Flush()
	//}
	time.Sleep(1 * time.Hour)
}
