// Package gg implements ginger graph creation, traversal, and (de)serialization
package gg

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"hash"
)

// TODO rename half-edge to open-edge

// Identifier is implemented by any value which can return a unique string for
// itself via an Identify method
type Identifier interface {
	Identify(hash.Hash)
}

func identify(i Identifier) string {
	h := md5.New()
	i.Identify(h)
	return hex.EncodeToString(h.Sum(nil))
}

// VertexType enumerates the different possible vertex types
type VertexType string

const (
	// Value is a Vertex which contains exactly one value and has at least one
	// edge (either input or output)
	Value VertexType = "value"

	// Junction is a Vertex which contains two or more in edges and exactly one
	// out edge
	Junction VertexType = "junction"
)

// Edge is a uni-directional connection between two vertices with an attribute
// value
type Edge struct {
	From  *Vertex
	Value Identifier
	To    *Vertex
}

// Vertex is a vertex in a Graph. No fields should be modified directly, only
// through method calls
type Vertex struct {
	VertexType
	Value   Identifier // Value is valid if-and-only-if VertexType is Value
	In, Out []Edge
}

////////////////////////////////////////////////////////////////////////////////

// HalfEdge is an un-realized Edge which can't be used for anything except
// constructing graphs. It has no meaning on its own.
type HalfEdge struct {
	// fromV will be the source vertex as-if the vertex (and any sub-vertices of
	// it) doesn't already exist in the graph. If it or it's sub-vertices does
	// already that will need to be taken into account when persisting into the
	// graph
	fromV vertex
	val   Identifier
}

// Identify implements the Identifier interface
func (he HalfEdge) Identify(h hash.Hash) {
	fmt.Fprintln(h, "halfEdge")
	he.fromV.Identify(h)
	he.val.Identify(h)
}

// vertex is a representation of a vertex in the graph. Each Graph contains a
// set of all the Value vertex instances it knows about. Each of these contains
// all the input HalfEdges which are known for it. So you can think of these
// "top-level" Value vertex instances as root nodes in a tree, and each HalfEdge
// as a branch.
//
// If a HalfEdge contains a fromV which is a Value that vertex won't have its in
// slice populated no matter what. If fromV is a Junction it will be populated,
// with any sub-Value's not being populated and so-on recursively
//
// When a view is constructed in makeView these Value instances are deduplicated
// and the top-level one's in value is used to properly connect it.
type vertex struct {
	VertexType
	val Identifier
	in  []HalfEdge
}

// A Value vertex is unique by the value it contains
// A Junction vertex is unique by its input edges
func (v vertex) Identify(h hash.Hash) {
	switch v.VertexType {
	case Value:
		fmt.Fprintln(h, "value")
		v.val.Identify(h)
	case Junction:
		fmt.Fprintf(h, "junction:%d\n", len(v.in))
		for _, in := range v.in {
			in.Identify(h)
		}
	default:
		panic(fmt.Sprintf("invalid VertexType:%#v", v))
	}
}

func (v vertex) cp() vertex {
	cp := v
	cp.in = make([]HalfEdge, len(v.in))
	copy(cp.in, v.in)
	return cp
}

func (v vertex) hasHalfEdge(he HalfEdge) bool {
	heID := identify(he)
	for _, in := range v.in {
		if identify(in) == heID {
			return true
		}
	}
	return false
}

func (v vertex) cpAndDelHalfEdge(he HalfEdge) (vertex, bool) {
	heID := identify(he)
	for i, in := range v.in {
		if identify(in) == heID {
			v = v.cp()
			v.in = append(v.in[:i], v.in[i+1:]...)
			return v, true
		}
	}
	return v, false
}

// Graph is a wrapper around a set of connected Vertices
type Graph struct {
	vM   map[string]vertex // only contains value vertices
	view map[string]*Vertex
}

// Null is the root empty graph, and is the base off which all graphs are built
var Null = &Graph{
	vM:   map[string]vertex{},
	view: map[string]*Vertex{},
}

// this does _not_ copy the view, as it's assumed the only reason to copy a
// graph is to modify it anyway
func (g *Graph) cp() *Graph {
	cp := &Graph{
		vM: make(map[string]vertex, len(g.vM)),
	}
	for id, v := range g.vM {
		cp.vM[id] = v
	}
	return cp
}

////////////////////////////////////////////////////////////////////////////////
// Graph creation

// ValueOut creates a HalfEdge which, when used to construct a Graph, represents
// an edge (with edgeVal attached to it) leaving the Value Vertex containing
// val.
//
// When constructing Graphs Value vertices are de-duplicated on their value. So
// multiple ValueOut HalfEdges constructed with the same val will be leaving the
// same Vertex instance in the constructed Graph.
func ValueOut(val, edgeVal Identifier) HalfEdge {
	return HalfEdge{
		fromV: vertex{
			VertexType: Value,
			val:        val,
		},
		val: edgeVal,
	}
}

// JunctionOut creates a HalfEdge which, when used to construct a Graph,
// represents an edge (with edgeVal attached to it) leaving the Junction Vertex
// comprised of the given ordered-set of input edges.
//
// When constructing Graphs Junction vertices are de-duplicated on their input
// edges. So multiple Junction HalfEdges constructed with the same set of input
// edges will be leaving the same Junction instance in the constructed Graph.
func JunctionOut(in []HalfEdge, edgeVal Identifier) HalfEdge {
	return HalfEdge{
		fromV: vertex{
			VertexType: Junction,
			in:         in,
		},
		val: edgeVal,
	}
}

// AddValueIn takes a HalfEdge and connects it to the Value Vertex containing
// val, returning the new Graph which reflects that connection. Any Vertices
// referenced within the HalfEdge which do not yet exist in the Graph will also
// be created in this step.
func (g *Graph) AddValueIn(he HalfEdge, val Identifier) *Graph {
	to := vertex{
		VertexType: Value,
		val:        val,
	}
	toID := identify(to)

	// if to is already in the graph, pull it out, as it might have existing in
	// edges we want to keep
	if exTo, ok := g.vM[toID]; ok {
		to = exTo
	}

	// if the incoming edge already exists in to then there's nothing to do
	if to.hasHalfEdge(he) {
		return g
	}

	to = to.cp()
	to.in = append(to.in, he)
	g = g.cp()

	// starting with to (which we always overwrite) go through vM and
	// recursively add in any vertices which aren't already there
	var persist func(vertex)
	persist = func(v vertex) {
		vID := identify(v)
		if v.VertexType == Value {
			if _, ok := g.vM[vID]; !ok {
				g.vM[vID] = v
			}
		} else {
			for _, e := range v.in {
				persist(e.fromV)
			}
		}
	}
	delete(g.vM, toID)
	persist(to)
	for _, e := range to.in {
		persist(e.fromV)
	}

	return g
}

// DelValueIn takes a HalfEdge and disconnects it from the Value Vertex
// containing val, returning the new Graph which reflects the disconnection. If
// the Value Vertex doesn't exist within the graph, or it doesn't have the given
// HalfEdge, no changes are made. Any vertices referenced by the HalfEdge for
// which that edge is their only outgoing edge will be removed from the Graph.
func (g *Graph) DelValueIn(he HalfEdge, val Identifier) *Graph {
	to := vertex{
		VertexType: Value,
		val:        val,
	}
	toID := identify(to)

	// pull to out of the graph. if it's not there then bail
	var ok bool
	if to, ok = g.vM[toID]; !ok {
		return g
	}

	// get new copy of to without the half-edge, or return if the half-edge
	// wasn't even in to
	to, ok = to.cpAndDelHalfEdge(he)
	if !ok {
		return g
	}
	g = g.cp()
	g.vM[toID] = to

	// connectedTo returns whether the vertex has any connections with the
	// vertex of the given id, descending recursively
	var connectedTo func(string, vertex) bool
	connectedTo = func(vID string, curr vertex) bool {
		for _, in := range curr.in {
			if in.fromV.VertexType == Value && identify(in.fromV) == vID {
				return true
			} else if in.fromV.VertexType == Junction && connectedTo(vID, in.fromV) {
				return true
			}
		}
		return false
	}

	// isOrphaned returns whether the given vertex has any connections to other
	// nodes in the graph
	isOrphaned := func(v vertex) bool {
		vID := identify(v)
		if v, ok := g.vM[vID]; ok && len(v.in) > 0 {
			return false
		}
		for vID2, v2 := range g.vM {
			if vID2 == vID {
				continue
			} else if connectedTo(vID, v2) {
				return false
			}
		}
		return true
	}

	// if to is orphaned get rid of it
	if isOrphaned(to) {
		delete(g.vM, toID)
	}

	// rmOrphaned descends down the given HalfEdge and removes any Value
	// Vertices referenced in it which are now orphaned
	var rmOrphaned func(HalfEdge)
	rmOrphaned = func(he HalfEdge) {
		if he.fromV.VertexType == Value && isOrphaned(he.fromV) {
			delete(g.vM, identify(he.fromV))
		} else if he.fromV.VertexType == Junction {
			for _, juncHe := range he.fromV.in {
				rmOrphaned(juncHe)
			}
		}
	}
	rmOrphaned(he)

	return g
}

// Union takes in another Graph and returns a new one which is the union of the
// two. Value vertices which are shared between the two will be merged so that
// the new vertex has the input edges of both.
func (g *Graph) Union(g2 *Graph) *Graph {
	g = g.cp()
	for vID, v2 := range g2.vM {
		v, ok := g.vM[vID]
		if !ok {
			v = v2
		} else {
			for _, v2e := range v2.in {
				if !v.hasHalfEdge(v2e) {
					v.in = append(v.in, v2e)
				}
			}
		}
		g.vM[vID] = v
	}
	return g
}

////////////////////////////////////////////////////////////////////////////////
// Graph traversal

func (g *Graph) makeView() {
	if g.view != nil {
		return
	}

	// view only contains value vertices, but we need to keep track of all
	// vertices while constructing the view
	g.view = make(map[string]*Vertex, len(g.vM))
	all := map[string]*Vertex{}

	var getV func(vertex, bool) *Vertex
	getV = func(v vertex, top bool) *Vertex {
		vID := identify(v)
		V, ok := all[vID]
		if !ok {
			V = &Vertex{VertexType: v.VertexType, Value: v.val}
			all[vID] = V
		}

		// we can be sure all Value vertices will be called with top==true at
		// some point, so we only need to descend into the input edges if:
		// * top is true
		// * this is a junction's first time being gotten
		if !top && (ok || v.VertexType != Junction) {
			return V
		}

		V.In = make([]Edge, 0, len(v.in))
		for i := range v.in {
			fromV := getV(v.in[i].fromV, false)
			e := Edge{From: fromV, Value: v.in[i].val, To: V}
			fromV.Out = append(fromV.Out, e)
			V.In = append(V.In, e)
		}

		if v.VertexType == Value {
			g.view[identify(v.val)] = V
		}

		return V
	}

	for _, v := range g.vM {
		getV(v, true)
	}
}

// Value returns the Value Vertex for the given value. If the Graph doesn't
// contain a vertex for the value then nil is returned
func (g *Graph) Value(val Identifier) *Vertex {
	g.makeView()
	return g.view[identify(val)]
}

// Values returns all Value Vertices in the Graph
func (g *Graph) Values() []*Vertex {
	g.makeView()
	vv := make([]*Vertex, 0, len(g.view))
	for _, v := range g.view {
		vv = append(vv, v)
	}
	return vv
}

// Equal returns whether or not the two Graphs are equivalent in value
func Equal(g1, g2 *Graph) bool {
	if len(g1.vM) != len(g2.vM) {
		return false
	}
	for v1ID, v1 := range g1.vM {
		v2, ok := g2.vM[v1ID]
		if !ok {
			return false
		}

		// since the vertices are values we must make sure their input sets are
		// the same (which is tricky since they're unordered, unlike a
		// junction's)
		if len(v1.in) != len(v2.in) {
			return false
		}
		for _, in := range v1.in {
			if !v2.hasHalfEdge(in) {
				return false
			}
		}
	}
	return true
}