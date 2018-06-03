package main

import (
	"sort"

	"github.com/mediocregopher/ginger/gg"
	"github.com/mediocregopher/ginger/gim/constraint"
	"github.com/mediocregopher/ginger/gim/geo"
	"github.com/mediocregopher/ginger/gim/terminal"
)

// "Solves" vertex position by detemining relative positions of vertices in
// primary and secondary directions (independently), with relative positions
// being described by "levels", where multiple vertices can occupy one level.
//
// Primary determines relative position in the primary direction by trying
// to place vertices before their outs and after their ins.
//
// Secondary determines relative position in the secondary direction by
// trying to place vertices relative to vertices they share an edge with in
// the order that the edges appear on the shared node.
func posSolve(g *gg.Graph) ([][]*gg.Vertex, map[string]int, map[string]int) {
	primEng := constraint.NewEngine()
	secEng := constraint.NewEngine()

	strM := g.ByID()
	for _, v := range strM {
		var prevIn *gg.Vertex
		for _, e := range v.In {
			primEng.AddConstraint(constraint.Constraint{
				Elem: e.From.ID,
				LT:   v.ID,
			})
			if prevIn != nil {
				secEng.AddConstraint(constraint.Constraint{
					Elem: prevIn.ID,
					LT:   e.From.ID,
				})
			}
			prevIn = e.From
		}

		var prevOut *gg.Vertex
		for _, e := range v.Out {
			if prevOut == nil {
				continue
			}
			secEng.AddConstraint(constraint.Constraint{
				Elem: prevOut.ID,
				LT:   e.To.ID,
			})
			prevOut = e.To
		}
	}
	prim := primEng.Solve()
	sec := secEng.Solve()

	// determine maximum primary level
	var maxPrim int
	for _, lvl := range prim {
		if lvl > maxPrim {
			maxPrim = lvl
		}
	}

	outStr := make([][]string, maxPrim+1)
	for v, lvl := range prim {
		outStr[lvl] = append(outStr[lvl], v)
	}

	// sort each primary level
	for _, vv := range outStr {
		sort.Slice(vv, func(i, j int) bool {
			return sec[vv[i]] < sec[vv[j]]
		})
	}

	// convert to vertices
	out := make([][]*gg.Vertex, len(outStr))
	for i, vv := range outStr {
		out[i] = make([]*gg.Vertex, len(outStr[i]))
		for j, v := range vv {
			out[i][j] = strM[v]
		}
	}
	return out, prim, sec
}

// mutates the boxes to be centered around the given point, keeping their
// relative position to each other
func centerBoxes(boxes []*box, around geo.XY) {
	var graphRect geo.Rect
	for _, b := range boxes {
		graphRect = graphRect.Union(b.rect())
	}
	graphMid := graphRect.Center(rounder)
	delta := around.Sub(graphMid)
	for _, b := range boxes {
		b.topLeft = b.topLeft.Add(delta)
	}
}

type view struct {
	g                       *gg.Graph
	primFlowDir, secFlowDir geo.XY
	start                   gg.Value
	center                  geo.XY // TODO this shouldnt be needed
}

func (view *view) draw(term *terminal.Terminal) {
	relPos, _, secSol := posSolve(view.g)

	// create boxes
	var boxes []*box
	boxesM := map[*box]*gg.Vertex{}
	boxesMr := map[*gg.Vertex]*box{}
	const (
		primPadding = 5
		secPadding  = 1
	)
	var primPos int
	for _, vv := range relPos {
		var primBoxes []*box // boxes on just this level
		var maxPrim int
		var secPos int
		for _, v := range vv {
			primVec := view.primFlowDir.Scale(primPos)
			secVec := view.secFlowDir.Scale(secPos)

			b := boxFromVertex(v, view.primFlowDir)
			b.topLeft = primVec.Add(secVec)
			boxes = append(boxes, &b)
			primBoxes = append(primBoxes, &b)
			boxesM[&b] = v
			boxesMr[v] = &b

			bSize := b.rect().Size
			primBoxLen := bSize.Mul(view.primFlowDir).Len(rounder)
			secBoxLen := bSize.Mul(view.secFlowDir).Len(rounder)
			if primBoxLen > maxPrim {
				maxPrim = primBoxLen
			}
			secPos += secBoxLen + secPadding
		}
		centerBoxes(primBoxes, view.primFlowDir.Scale(primPos))
		primPos += maxPrim + primPadding
	}

	// maps a vertex to all of its to edges, sorted by secSol
	findFromIM := map[*gg.Vertex][]gg.Edge{}
	// returns the index of this edge in from's Out
	findFromI := func(from *gg.Vertex, e gg.Edge) int {
		edges, ok := findFromIM[from]
		if !ok {
			edges = make([]gg.Edge, len(from.Out))
			copy(edges, from.Out)
			sort.Slice(edges, func(i, j int) bool {
				// TODO if two edges go to the same vertex, how are they sorted?
				return secSol[edges[i].To.ID] < secSol[edges[j].To.ID]
			})
			findFromIM[from] = edges
		}

		for i, fe := range edges {
			if fe == e {
				return i
			}
		}
		panic("edge not found in from.Out")
	}

	// create lines
	var lines []line
	for _, b := range boxes {
		v := boxesM[b]
		for i, e := range v.In {
			bFrom := boxesMr[e.From]
			fromI := findFromI(e.From, e)
			lines = append(lines, line{
				from:  bFrom,
				fromI: fromI,
				to:    b,
				toI:   i,
				body:  e.Value.V.(string),
			})
		}
	}

	// translate all boxes so the graph is centered around v.center
	centerBoxes(boxes, view.center)

	// actually draw the boxes and lines
	buf := terminal.NewBuffer()
	for _, b := range boxes {
		b.draw(buf)
	}
	for _, line := range lines {
		line.draw(buf, view.primFlowDir, view.secFlowDir)
	}
	term.WriteBuffer(geo.Zero, buf)
}
