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
// - actually use flowDir
// - assign edges to "slots" on boxes
// - figure out how to keep boxes sorted on their levels (e.g. the "b" nodes)

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
	termSize := term.WindowSize()
	g := mkGraph()

	// level 0 is at the bottom of the screen, cause life is easier that way
	levels := map[*gg.Vertex]int{}
	getLevel := func(v *gg.Vertex) int {
		// if any of the tos have a level, this will be greater than the max
		toMax := -1
		for _, e := range v.Out {
			lvl, ok := levels[e.To]
			if !ok {
				continue
			} else if lvl > toMax {
				toMax = lvl
			}
		}

		if toMax >= 0 {
			return toMax + 1
		}

		// otherwise level is 0
		return 0
	}

	g.Walk(g.Value(str("c")), func(v *gg.Vertex) bool {
		levels[v] = getLevel(v)
		return true
	})

	// consolidate by level
	byLevel := map[int][]*gg.Vertex{}
	maxLvl := -1
	for v, lvl := range levels {
		byLevel[lvl] = append(byLevel[lvl], v)
		if lvl > maxLvl {
			maxLvl = lvl
		}
	}

	// create boxes
	boxes := map[*gg.Vertex]box{}
	for lvl := 0; lvl <= maxLvl; lvl++ {
		vv := byLevel[lvl]
		for i, v := range vv {
			b := boxFromVertex(v, geo.Right)
			bSize := b.rect().Size
			b.topLeft = geo.XY{
				10*(i-(len(vv)/2)) - (bSize[0] / 2),
				lvl * -10,
			}
			boxes[v] = b
		}
	}

	// center boxes. first find overall dimensions, use that to create delta
	// vector which would move that to the center
	var graphRect geo.Rect
	for _, b := range boxes {
		graphRect = graphRect.Union(b.rect())
	}

	graphMid := graphRect.Center(rounder)
	screenMid := geo.Zero.Midpoint(termSize, rounder)
	delta := screenMid.Sub(graphMid)

	// translate all boxes by delta
	for v, b := range boxes {
		b.topLeft = b.topLeft.Add(delta)
		boxes[v] = b
	}

	// create lines
	var lines [][2]box
	for v := range levels {
		b := boxes[v]
		for _, e := range v.In {
			bFrom := boxes[e.From]
			lines = append(lines, [2]box{bFrom, b})
		}
	}

	for range time.Tick(frameperiod) {
		// update phase
		// nufin

		// draw phase
		term.Reset()
		for v := range boxes {
			boxes[v].draw(term)
		}
		for _, line := range lines {
			basicLine(term, line[0], line[1])
		}
		term.Flush()
	}
}
