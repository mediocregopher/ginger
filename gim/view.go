package main

import (
	"github.com/mediocregopher/ginger/gg"
	"github.com/mediocregopher/ginger/gim/geo"
	"github.com/mediocregopher/ginger/gim/terminal"
)

type view struct {
	g       *gg.Graph
	flowDir geo.XY
	start   str
	center  geo.XY
}

func (v *view) draw(term *terminal.Terminal) {
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

	v.g.Walk(v.g.Value(v.start), func(v *gg.Vertex) bool {
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
	boxes := map[*gg.Vertex]*box{}
	for lvl := 0; lvl <= maxLvl; lvl++ {
		vv := byLevel[lvl]
		for i, v := range vv {
			b := boxFromVertex(v, geo.Right)
			bSize := b.rect().Size
			// TODO make this dependent on flowDir
			b.topLeft = geo.XY{
				10*(i-(len(vv)/2)) - (bSize[0] / 2),
				lvl * -10,
			}
			boxes[v] = &b
		}
	}

	// create lines
	var lines []line
	for v := range levels {
		b := boxes[v]
		for _, e := range v.In {
			bFrom := boxes[e.From]
			lines = append(lines, line{bFrom, b})
		}
	}

	// translate all boxes so the graph is centered around v.center. Since the
	// lines use pointers to the boxes this will update them as well
	var graphRect geo.Rect
	for _, b := range boxes {
		graphRect = graphRect.Union(b.rect())
	}
	graphMid := graphRect.Center(rounder)
	delta := v.center.Sub(graphMid)
	for _, b := range boxes {
		b.topLeft = b.topLeft.Add(delta)
	}

	// actually draw the boxes and lines
	for _, box := range boxes {
		box.draw(term)
	}
	for _, line := range lines {
		line.draw(term, v.flowDir)
	}
}
