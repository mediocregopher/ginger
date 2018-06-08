// Package view implements rendering a graph to a terminal.
//
// Steps for rendering
//
//	- Preprocessing: Disjoin Graph into multiple Graphs, and decide how to
//	  arrange them (maybe sort by number of vertices or number of edges (or the
//	  sum of both) or something).
//
//	- Convert Graph into internal representation.
//		- Still uses gg.Graph, but vertices and edge values are wrapped in types
//		  internal to this package, and on which further mapping will be done.
//		- Positions unknown at this point.
//		- Junctions are converted to value vertices with set edge order.
//		- Edges contain both their body and their tail/head rune.
//
//	- Find eligible "root" vertex, probably by one which has the fewest input
//	  edges.
//
//	- Find cycles and reverse edges as needed.
//		- The to/from vertices are reversed, as are the head/tail runes, so the
//		  direction will appear consistent with the original graph
//		- TODO this might not be necessary? Or at least may need to be modified.
//		  In the paper this is done, but that algorithm allows for edges upward
//		  from their tail, whereas this one doesn't. It might only be necessary
//		  for the MST stuff, in which case this might only need to take place
//		  within Positioning-Part1.
//
//	- Replace edge bodies with a vertex with a single input/output edge.
//
//	- Position all vertices
//		- `coord` field on vertices used as row/column coordinates.
//		- Positioning will be done with down being the primary direction and
//		  right being the secondary direction.
//		- Part 1) find vertical positions for all vertices (aka assign rows)
//			- This step uses some fancy MST stuff as outlined by (TODO refer to
//			  paper here).
//		- Part 2) find horizontal positions within rows (aka assign columns)
//			- Part of this will include creating ephemeral vertices where an
//			  edge spans a row without having a vertex on it. These will be
//			  removed as the final part of this step.
//			- The jist of this step is to find vertex ordering which reduces
//			  number of edge crossings between adjacent rows.
//			- Some extra care is taken for cases where an edge's from vertex is
//			  not a lower row than its to vertex.
//				- This is an unavoidable case, as at the least a vertex may
//				  connect to itself.
//				- These edges will have their `switchback` field set to true.
//				- For the purposes of calculating edge crossings these edges
//				  should be ignored. During the absolute positioning and drawing
//				  steps they will be accounted for and dealt with.
//		- Part 3) row/column positions into terminal positions, which are
//		  stored on the vertices in the `pos` field. Primary/secondary
//		  direction are taken into account here.
//
//	- Post-processing: any additional absolute positioning and other formatting
//	  given by the user for the Graph should be done here
//
//	- Draw vertices and their edges to buffer
//		- At this point drawing vertices is easy. Edges is more complicated but
//		  the start/end positions of each edge should already be known, so while
//		  drawing may be complex it's not difficult.
//
package view

import (
	"sort"

	"github.com/mediocregopher/ginger/gg"
	"github.com/mediocregopher/ginger/gim/geo"
	"github.com/mediocregopher/ginger/gim/terminal"
	"github.com/mediocregopher/ginger/gim/view/constraint"
)

// View wraps a single Graph instance and a set of display options for it, and
// generates renderable terminal output for it.
type View struct {
	g     *gg.Graph
	start gg.Value // TODO shouldn't need this

	primFlowDir, secFlowDir geo.XY
}

// New instantiates and returns a view around the given Graph instance, with
// start indicating the value vertex to consider the "root" of the graph.
//
// Drawing is done by aligning the vertices into rows and columns in such a way
// as to reduce edge crossings. primaryDir indicates the direction edges will
// primarily be pointed in. For example, if it is geo.Down then adjacent
// vertices will be arranged into columns.
//
// secondaryDir indicates the direction vertices should be arranged when they
// end up in the same "rank" (e.g. when primaryDir is geo.Down, all vertices on
// the same row will be the same "rank").
//
// A primaryDir/secondaryDir of either geo.Down/geo.Right or geo.Right/geo.Down
// are recommended, but any combination of perpendicular directions is allowed.
func New(g *gg.Graph, start gg.Value, primaryDir, secondaryDir geo.XY) *View {
	return &View{
		g:           g,
		start:       start,
		primFlowDir: primaryDir,
		secFlowDir:  secondaryDir,
	}
}

// Draw renders and draws the View's Graph to the Buffer.
func (view *View) Draw(buf *terminal.Buffer) {
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
			primBoxLen := bSize.Mul(view.primFlowDir).Len()
			secBoxLen := bSize.Mul(view.secFlowDir).Len()
			if primBoxLen > maxPrim {
				maxPrim = primBoxLen
			}
			secPos += secBoxLen + secPadding
		}
		for _, b := range primBoxes {
			b.topLeft = b.topLeft.Add(view.primFlowDir.Scale(primPos))
		}
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
			buf := terminal.NewBuffer()
			buf.WriteString(e.Value.V.(string))
			lines = append(lines, line{
				from:    bFrom,
				fromI:   fromI,
				to:      b,
				toI:     i,
				bodyBuf: buf,
			})
		}
	}

	// actually draw the boxes and lines
	for _, b := range boxes {
		b.draw(buf)
	}
	for _, line := range lines {
		line.draw(buf, view.primFlowDir, view.secFlowDir)
	}
}

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
