// Package gg implements ginger graph creation, traversal, and (de)serialization
package gg

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
)

// Value wraps a go value in a way such that it will be uniquely identified
// within any Graph and between Graphs. Use NewValue to create a Value instance.
// You can create an instance manually as long as ID is globally unique.
type Value struct {
	ID string
	V  interface{}
}

// NewValue returns a Value instance wrapping any go value. The Value returned
// will be independent of the passed in go value. So if the same go value is
// passed in twice then the two returned Value instances will be treated as
// being different values by Graph.
func NewValue(V interface{}) Value {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return Value{
		ID: hex.EncodeToString(b),
		V:  V,
	}
}

// VertexType enumerates the different possible vertex types
type VertexType string

const (
	// ValueVertex is a Vertex which contains exactly one value and has at least
	// one edge (either input or output)
	ValueVertex VertexType = "value"

	// JunctionVertex is a Vertex which contains two or more in edges and
	// exactly one out edge
	JunctionVertex VertexType = "junction"
)

// Edge is a uni-directional connection between two vertices with an attribute
// value
type Edge struct {
	From  *Vertex
	Value Value
	To    *Vertex
}

// Vertex is a vertex in a Graph. No fields should be modified directly, only
// through method calls
type Vertex struct {
	ID string
	VertexType
	Value   Value // Value is valid if-and-only-if VertexType is ValueVertex
	In, Out []Edge
}

////////////////////////////////////////////////////////////////////////////////

// OpenEdge is an un-realized Edge which can't be used for anything except
// constructing graphs. It has no meaning on its own.
type OpenEdge struct {
	// fromV will be the source vertex as-if the vertex (and any sub-vertices of
	// it) doesn't already exist in the graph. If it or it's sub-vertices does
	// already that will need to be taken into account when persisting into the
	// graph
	fromV vertex
	val   Value
}

func (oe OpenEdge) id() string {
	return fmt.Sprintf("(%s,%s)", oe.fromV.id, oe.val.ID)
}

// vertex is a representation of a vertex in the graph. Each Graph contains a
// set of all the Value vertex instances it knows about. Each of these contains
// all the input OpenEdges which are known for it. So you can think of these
// "top-level" Value vertex instances as root nodes in a tree, and each OpenEdge
// as a branch.
//
// If a OpenEdge contains a fromV which is a Value that vertex won't have its in
// slice populated no matter what. If fromV is a Junction it will be populated,
// with any sub-Value's not being populated and so-on recursively
//
// When a view is constructed in makeView these Value instances are deduplicated
// and the top-level one's in value is used to properly connect it.
type vertex struct {
	id string
	VertexType
	val Value
	in  []OpenEdge
}

func (v vertex) cp() vertex {
	cp := v
	cp.in = make([]OpenEdge, len(v.in))
	copy(cp.in, v.in)
	return cp
}

func (v vertex) hasOpenEdge(oe OpenEdge) bool {
	oeID := oe.id()
	for _, in := range v.in {
		if in.id() == oeID {
			return true
		}
	}
	return false
}

func (v vertex) cpAndDelOpenEdge(oe OpenEdge) (vertex, bool) {
	oeID := oe.id()
	for i, in := range v.in {
		if in.id() == oeID {
			v = v.cp()
			v.in = append(v.in[:i], v.in[i+1:]...)
			return v, true
		}
	}
	return v, false
}

// Graph is a wrapper around a set of connected Vertices
type Graph struct {
	vM map[string]vertex // only contains value vertices

	// generated by makeView on-demand
	byVal map[string]*Vertex
	all   map[string]*Vertex
}

// Null is the root empty graph, and is the base off which all graphs are built
var Null = &Graph{
	vM:    map[string]vertex{},
	byVal: map[string]*Vertex{},
	all:   map[string]*Vertex{},
}

// this does _not_ copy the view, as it's assumed the only reason to copy a
// graph is to modify it anyway
func (g *Graph) cp() *Graph {
	cp := &Graph{
		vM: make(map[string]vertex, len(g.vM)),
	}
	for vID, v := range g.vM {
		cp.vM[vID] = v
	}
	return cp
}

////////////////////////////////////////////////////////////////////////////////
// Graph creation

func mkVertex(typ VertexType, val Value, ins []OpenEdge) vertex {
	v := vertex{VertexType: typ, in: ins}
	switch typ {
	case ValueVertex:
		v.id = val.ID
		v.val = val
	case JunctionVertex:
		inIDs := make([]string, len(ins))
		for i := range ins {
			inIDs[i] = ins[i].id()
		}
		v.id = "[" + strings.Join(inIDs, ",") + "]"
	default:
		panic(fmt.Sprintf("unknown vertex type %q", typ))
	}
	return v
}

// ValueOut creates a OpenEdge which, when used to construct a Graph, represents
// an edge (with edgeVal attached to it) coming from the ValueVertex containing
// val.
//
// When constructing Graphs, Value vertices are de-duplicated on their Value. So
// multiple ValueOut OpenEdges constructed with the same val will be leaving the
// same Vertex instance in the constructed Graph.
func ValueOut(val, edgeVal Value) OpenEdge {
	return OpenEdge{fromV: mkVertex(ValueVertex, val, nil), val: edgeVal}
}

// JunctionOut creates a OpenEdge which, when used to construct a Graph,
// represents an edge (with edgeVal attached to it) coming from the
// JunctionVertex comprised of the given ordered-set of input edges.
//
// When constructing Graphs Junction vertices are de-duplicated on their input
// edges. So multiple Junction OpenEdges constructed with the same set of input
// edges will be leaving the same Junction instance in the constructed Graph.
func JunctionOut(in []OpenEdge, edgeVal Value) OpenEdge {
	return OpenEdge{
		fromV: mkVertex(JunctionVertex, Value{}, in),
		val:   edgeVal,
	}
}

// AddValueIn takes a OpenEdge and connects it to the Value Vertex containing
// val, returning the new Graph which reflects that connection. Any Vertices
// referenced within toe OpenEdge which do not yet exist in the Graph will also
// be created in this step.
func (g *Graph) AddValueIn(oe OpenEdge, val Value) *Graph {
	to := mkVertex(ValueVertex, val, nil)
	toID := to.id

	// if to is already in the graph, pull it out, as it might have existing in
	// edges we want to keep
	if exTo, ok := g.vM[toID]; ok {
		to = exTo
	}

	// if the incoming edge already exists in to then there's nothing to do
	if to.hasOpenEdge(oe) {
		return g
	}

	to = to.cp()
	to.in = append(to.in, oe)
	g = g.cp()

	// starting with to (which we always overwrite) go through vM and
	// recursively add in any vertices which aren't already there
	var persist func(vertex)
	persist = func(v vertex) {
		if v.VertexType == ValueVertex {
			vID := v.id
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

// DelValueIn takes a OpenEdge and disconnects it from the Value Vertex
// containing val, returning the new Graph which reflects the disconnection. If
// the Value Vertex doesn't exist within the graph, or it doesn't have the given
// OpenEdge, no changes are made. Any vertices referenced by toe OpenEdge for
// which that edge is their only outgoing edge will be removed from the Graph.
func (g *Graph) DelValueIn(oe OpenEdge, val Value) *Graph {
	to := mkVertex(ValueVertex, val, nil)
	toID := to.id

	// pull to out of the graph. if it's not there then bail
	var ok bool
	if to, ok = g.vM[toID]; !ok {
		return g
	}

	// get new copy of to without the half-edge, or return if the half-edge
	// wasn't even in to
	to, ok = to.cpAndDelOpenEdge(oe)
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
			if in.fromV.VertexType == ValueVertex && in.fromV.id == vID {
				return true
			} else if in.fromV.VertexType == JunctionVertex && connectedTo(vID, in.fromV) {
				return true
			}
		}
		return false
	}

	// isOrphaned returns whether the given vertex has any connections to other
	// nodes in the graph
	isOrphaned := func(v vertex) bool {
		vID := v.id
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

	// rmOrphaned descends down the given OpenEdge and removes any Value
	// Vertices referenced in it which are now orphaned
	var rmOrphaned func(OpenEdge)
	rmOrphaned = func(oe OpenEdge) {
		if oe.fromV.VertexType == ValueVertex && isOrphaned(oe.fromV) {
			delete(g.vM, oe.fromV.id)
		} else if oe.fromV.VertexType == JunctionVertex {
			for _, juncOe := range oe.fromV.in {
				rmOrphaned(juncOe)
			}
		}
	}
	rmOrphaned(oe)

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
				if !v.hasOpenEdge(v2e) {
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
	if g.byVal != nil {
		return
	}

	g.byVal = make(map[string]*Vertex, len(g.vM))
	g.all = map[string]*Vertex{}

	var getV func(vertex, bool) *Vertex
	getV = func(v vertex, top bool) *Vertex {
		V, ok := g.all[v.id]
		if !ok {
			V = &Vertex{ID: v.id, VertexType: v.VertexType, Value: v.val}
			g.all[v.id] = V
		}

		// we can be sure all Value vertices will be called with top==true at
		// some point, so we only need to descend into the input edges if:
		// * top is true
		// * this is a junction's first time being gotten
		if !top && (ok || v.VertexType != JunctionVertex) {
			return V
		}

		V.In = make([]Edge, 0, len(v.in))
		for i := range v.in {
			fromV := getV(v.in[i].fromV, false)
			e := Edge{From: fromV, Value: v.in[i].val, To: V}
			fromV.Out = append(fromV.Out, e)
			V.In = append(V.In, e)
		}

		if v.VertexType == ValueVertex {
			g.byVal[v.val.ID] = V
		}

		return V
	}

	for _, v := range g.vM {
		getV(v, true)
	}
}

// ValueVertex returns the Value Vertex for the given value. If the Graph
// doesn't contain a vertex for the value then nil is returned
func (g *Graph) ValueVertex(val Value) *Vertex {
	g.makeView()
	return g.byVal[val.ID]
}

// ValueVertices returns all Value Vertices in the Graph
func (g *Graph) ValueVertices() []*Vertex {
	g.makeView()
	vv := make([]*Vertex, 0, len(g.byVal))
	for _, v := range g.byVal {
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
			if !v2.hasOpenEdge(in) {
				return false
			}
		}
	}
	return true
}

// TODO Walk, but by edge
// TODO Walk, but without end. AKA FSM

// Walk will traverse the Graph, calling the callback on every Vertex in the
// Graph once. If startWith is non-nil then that Vertex will be the first one
// passed to the callback and used as the starting point of the traversal. If
// the callback returns false the traversal is stopped.
func (g *Graph) Walk(startWith *Vertex, callback func(*Vertex) bool) {
	// TODO figure out how to make Walk deterministic?
	g.makeView()
	if len(g.byVal) == 0 {
		return
	}

	seen := make(map[*Vertex]bool, len(g.byVal))
	var innerWalk func(*Vertex) bool
	innerWalk = func(v *Vertex) bool {
		if seen[v] {
			return true
		} else if !callback(v) {
			return false
		}
		seen[v] = true
		for _, e := range v.In {
			if !innerWalk(e.From) {
				return false
			}
		}
		return true
	}

	if startWith != nil {
		if !innerWalk(startWith) {
			return
		}
	}

	for _, v := range g.byVal {
		if !innerWalk(v) {
			return
		}
	}
}

// ByID returns all vertices indexed by their ID field
func (g *Graph) ByID() map[string]*Vertex {
	g.makeView()
	return g.all
}
