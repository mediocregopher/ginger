// Package gg implements ginger graph creation, traversal, and (de)serialization
package gg

import (
	"fmt"
	"strings"
)

// Value represents a value being stored in a Graph. No more than one field may
// be non-nil. No fields being set indicates lack of value.
type Value struct {
	Name   *string
	Number *int64
	Graph  *Graph

	// TODO coming soon!
	// String *string
}

// Equal returns true if the passed in Value is equivalent.
func (v Value) Equal(v2 Value) bool {
	switch {

	case v == Value{} && v2 == Value{}:
		return true

	case v.Name != nil && v2.Name != nil && *v.Name == *v2.Name:
		return true

	case v.Number != nil && v2.Number != nil && *v.Number == *v2.Number:
		return true

	case v.Graph != nil && v2.Graph != nil && Equal(v.Graph, v2.Graph):
		return true

	default:
		return false
	}
}

func (v Value) String() string {

	switch {

	case v == Value{}:
		return "<noval>"

	case v.Name != nil:
		return *v.Name

	case v.Number != nil:
		return fmt.Sprint(*v.Number)

	case v.Graph != nil:
		return v.Graph.String()

	default:
		panic("unknown value kind")
	}
}

// VertexType enumerates the different possible vertex types.
type VertexType string

const (
	// ValueVertex is a Vertex which contains exactly one value and has at least
	// one edge (either input or output).
	ValueVertex VertexType = "val"

	// TupleVertex is a Vertex which contains two or more in edges and
	// exactly one out edge
	//
	// TODO ^ what about 0 or 1 in edges?
	TupleVertex VertexType = "tup"
)

////////////////////////////////////////////////////////////////////////////////

// OpenEdge is an un-realized Edge which can't be used for anything except
// constructing graphs. It has no meaning on its own.
type OpenEdge struct {
	fromV vertex
	val   Value
}

// WithEdgeVal returns a copy of the OpenEdge with the edge value replaced by
// the given one.
func (oe OpenEdge) WithEdgeVal(val Value) OpenEdge {
	oe.val = val
	return oe
}

func (oe OpenEdge) String() string {
	return fmt.Sprintf("%s(%s, %s)", oe.fromV.VertexType, oe.fromV.String(), oe.val.String())
}

// ValueOut creates a OpenEdge which, when used to construct a Graph, represents
// an edge (with edgeVal attached to it) coming from the ValueVertex containing
// val.
func ValueOut(val, edgeVal Value) OpenEdge {
	return OpenEdge{fromV: mkVertex(ValueVertex, val), val: edgeVal}
}

// TupleOut creates an OpenEdge which, when used to construct a Graph,
// represents an edge (with edgeVal attached to it) coming from the
// TupleVertex comprised of the given ordered-set of input edges.
//
// If len(ins) == 1 and edgeVal == Value{}, then that single OpenEdge is
// returned as-is.
func TupleOut(ins []OpenEdge, edgeVal Value) OpenEdge {

	if len(ins) == 1 {

		if edgeVal == (Value{}) {
			return ins[0]
		}

		if ins[0].val == (Value{}) {
			return ins[0].WithEdgeVal(edgeVal)
		}

	}

	return OpenEdge{
		fromV: mkVertex(TupleVertex, Value{}, ins...),
		val:   edgeVal,
	}
}

func (oe OpenEdge) equal(oe2 OpenEdge) bool {
	return oe.val.Equal(oe2.val) && oe.fromV.equal(oe2.fromV)
}

type vertex struct {
	VertexType
	val Value
	tup []OpenEdge
}

func mkVertex(typ VertexType, val Value, tupIns ...OpenEdge) vertex {
	return vertex{
		VertexType: typ,
		val:        val,
		tup:        tupIns,
	}
}

func (v vertex) equal(v2 vertex) bool {

	if v.VertexType != v2.VertexType {
		return false
	}

	if v.VertexType == ValueVertex {
		return v.val.Equal(v2.val)
	}

	if len(v.tup) != len(v2.tup) {
		return false
	}

	for i := range v.tup {
		if !v.tup[i].equal(v2.tup[i]) {
			return false
		}
	}

	return true
}

func (v vertex) String() string {

	switch v.VertexType {

	case ValueVertex:
		return v.val.String()

	case TupleVertex:

		strs := make([]string, len(v.tup))

		for i := range v.tup {
			strs[i] = v.tup[i].String()
		}

		return fmt.Sprintf("[%s]", strings.Join(strs, ", "))

	default:
		panic("unknown vertix kind")
	}

}

type graphValueIn struct {
	val   Value
	edges []OpenEdge
}

func (valIn graphValueIn) cp() graphValueIn {
	cp := valIn
	cp.edges = make([]OpenEdge, len(valIn.edges))
	copy(cp.edges, valIn.edges)
	return valIn
}

func (valIn graphValueIn) equal(valIn2 graphValueIn) bool {
	if !valIn.val.Equal(valIn2.val) {
		return false
	}

	if len(valIn.edges) != len(valIn2.edges) {
		return false
	}

outer:
	for _, edge := range valIn.edges {

		for _, edge2 := range valIn2.edges {

			if edge.equal(edge2) {
				continue outer
			}
		}

		return false
	}

	return true
}

// Graph is an immutable container of a set of vertices. The Graph keeps track
// of all Values which terminate an OpenEdge (which may be a tree of Value/Tuple
// vertices).
//
// NOTE The current implementation of Graph is incredibly inefficient, there's
// lots of O(N) operations, unnecessary copying on changes, and duplicate data
// in memory.
type Graph struct {
	valIns []graphValueIn
}

// ZeroGraph is the root empty graph, and is the base off which all graphs are
// built.
var ZeroGraph = &Graph{}

func (g *Graph) cp() *Graph {
	cp := &Graph{
		valIns: make([]graphValueIn, len(g.valIns)),
	}
	copy(cp.valIns, g.valIns)
	return cp
}

func (g *Graph) String() string {

	var strs []string

	for _, valIn := range g.valIns {
		for _, oe := range valIn.edges {
			strs = append(
				strs,
				fmt.Sprintf("valIn(%s, %s)", oe.String(), valIn.val.String()),
			)
		}
	}

	return fmt.Sprintf("graph(%s)", strings.Join(strs, ", "))
}

func (g *Graph) valIn(val Value) graphValueIn {
	for _, valIn := range g.valIns {
		if valIn.val.Equal(val) {
			return valIn
		}
	}

	return graphValueIn{val: val}
}

// AddValueIn takes a OpenEdge and connects it to the Value Vertex containing
// val, returning the new Graph which reflects that connection. Any Vertices
// referenced within toe OpenEdge which do not yet exist in the Graph will also
// be created in this step.
func (g *Graph) AddValueIn(oe OpenEdge, val Value) *Graph {

	valIn := g.valIn(val)

	for _, existingOE := range valIn.edges {
		if existingOE.equal(oe) {
			return g
		}
	}

	valIn = valIn.cp()
	valIn.edges = append(valIn.edges, oe)

	g = g.cp()

	for i, existingValIn := range g.valIns {
		if existingValIn.val.Equal(val) {
			g.valIns[i] = valIn
			return g
		}
	}

	g.valIns = append(g.valIns, valIn)
	return g
}

// Equal returns whether or not the two Graphs are equivalent in value.
func Equal(g1, g2 *Graph) bool {

	if len(g1.valIns) != len(g2.valIns) {
		return false
	}

outer:
	for _, valIn1 := range g1.valIns {

		for _, valIn2 := range g2.valIns {

			if valIn1.equal(valIn2) {
				continue outer
			}
		}

		return false
	}

	return true
}
