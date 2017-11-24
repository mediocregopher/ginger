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
// - assign edges to "slots" on boxes
// - finish initial implementation of constraint, use that to sort boxes by
//   primary and secondary flowDir based on their edges
// - be able to draw circular graphs

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

func mkGraph() *gg.Graph {
	aE0 := gg.ValueOut(gg.Str("a"), gg.Str("aE0"))
	aE1 := gg.ValueOut(gg.Str("a"), gg.Str("aE1"))
	aE2 := gg.ValueOut(gg.Str("a"), gg.Str("aE2"))
	aE3 := gg.ValueOut(gg.Str("a"), gg.Str("aE3"))
	g := gg.Null
	g = g.AddValueIn(aE0, gg.Str("b0"))
	g = g.AddValueIn(aE1, gg.Str("b1"))
	g = g.AddValueIn(aE2, gg.Str("b2"))
	g = g.AddValueIn(aE3, gg.Str("b3"))

	jE := gg.JunctionOut([]gg.OpenEdge{
		gg.ValueOut(gg.Str("b0"), gg.Str("")),
		gg.ValueOut(gg.Str("b1"), gg.Str("")),
		gg.ValueOut(gg.Str("b2"), gg.Str("")),
		gg.ValueOut(gg.Str("b3"), gg.Str("")),
	}, gg.Str("jE"))
	g = g.AddValueIn(jE, gg.Str("c"))
	return g
}

//func mkGraph() *gg.Graph {
//	g := gg.Null
//	g = g.AddValueIn(gg.ValueOut(str("a"), str("e")), str("b"))
//	return g
//}

func main() {
	rand.Seed(time.Now().UnixNano())
	term := terminal.New()
	term.Reset()
	term.HideCursor()

	v := view{
		g:       mkGraph(),
		flowDir: geo.Down,
		start:   gg.Str("c"),
		center:  geo.Zero.Midpoint(term.WindowSize(), rounder),
	}

	for range time.Tick(frameperiod) {
		term.Reset()
		v.draw(term)
		term.Flush()
	}
}
