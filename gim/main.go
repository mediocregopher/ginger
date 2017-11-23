package main

import (
	"fmt"
	"hash"
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
// - figure out how to keep boxes sorted on their levels (e.g. the "b" nodes)
// - be able to draw circular graphs

const (
	framerate   = 10
	frameperiod = time.Second / time.Duration(framerate)
	rounder     = geo.Ceil
)

type str string

func (s str) Identify(h hash.Hash) {
	fmt.Fprintln(h, s)
}

func debugf(str string, args ...interface{}) {
	if !strings.HasSuffix(str, "\n") {
		str += "\n"
	}
	fmt.Fprintf(os.Stderr, str, args...)
}

func mkGraph() *gg.Graph {
	aE0 := gg.ValueOut(str("a"), str("aE0"))
	aE1 := gg.ValueOut(str("a"), str("aE1"))
	aE2 := gg.ValueOut(str("a"), str("aE2"))
	aE3 := gg.ValueOut(str("a"), str("aE3"))
	g := gg.Null
	g = g.AddValueIn(aE0, str("b0"))
	g = g.AddValueIn(aE1, str("b1"))
	g = g.AddValueIn(aE2, str("b2"))
	g = g.AddValueIn(aE3, str("b3"))

	jE := gg.JunctionOut([]gg.OpenEdge{
		gg.ValueOut(str("b0"), str("")),
		gg.ValueOut(str("b1"), str("")),
		gg.ValueOut(str("b2"), str("")),
		gg.ValueOut(str("b3"), str("")),
	}, str("jE"))
	g = g.AddValueIn(jE, str("c"))
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
		start:   str("c"),
		center:  geo.Zero.Midpoint(term.WindowSize(), rounder),
	}

	for range time.Tick(frameperiod) {
		term.Reset()
		v.draw(term)
		term.Flush()
	}
}
