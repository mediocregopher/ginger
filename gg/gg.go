// Package gg implements ginger graph creation, traversal, and (de)serialization
package gg

import (
	"fmt"
	"strings"
)

// ZeroValue is a Value with no fields set.
var ZeroValue Value

// Value represents a value being stored in a Graph.
type Value struct {

	// Only one of these fields may be set
	Name   *string
	Number *int64
	Graph  *Graph

	// TODO coming soon!
	// String *string

	// Optional fields indicating the token which was used to construct this
	// Value, if any.
	LexerToken *LexerToken
}

// IsZero returns true if the Value is the zero value (none of the sub-value
// fields are set). LexerToken is ignored for this check.
func (v Value) IsZero() bool {
	v.LexerToken = nil
	return v == Value{}
}

// Equal returns true if the passed in Value is equivalent.
func (v Value) Equal(v2 Value) bool {
	switch {

	case v.IsZero() && v2.IsZero():
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

	case v.IsZero():
		return "<zero>"

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

////////////////////////////////////////////////////////////////////////////////

// OpenEdge is an un-realized Edge which can't be used for anything except
// constructing graphs. It has no meaning on its own.
type OpenEdge struct {
	fromV   vertex
	edgeVal Value
}

func (oe OpenEdge) String() string {

	vertexType := "tup"

	if oe.fromV.val != nil {
		vertexType = "val"
	}

	return fmt.Sprintf("%s(%s, %s)", vertexType, oe.fromV.String(), oe.edgeVal.String())
}

// EdgeValue returns the Value which lies on the edge itself.
func (oe OpenEdge) EdgeValue() Value {
	return oe.edgeVal
}

// FromValue returns the Value from which the OpenEdge was created via ValueOut,
// or false if it wasn't created via ValueOut.
func (oe OpenEdge) FromValue() (Value, bool) {
	if oe.fromV.val == nil {
		return ZeroValue, false
	}

	return *oe.fromV.val, true
}

// FromTuple returns the tuple of OpenEdges from which the OpenEdge was created
// via TupleOut, or false if it wasn't created via TupleOut.
func (oe OpenEdge) FromTuple() ([]OpenEdge, bool) {
	if oe.fromV.val != nil {
		return nil, false
	}

	return oe.fromV.tup, true
}

// ValueOut creates a OpenEdge which, when used to construct a Graph, represents
// an edge (with edgeVal attached to it) coming from the ValueVertex containing
// val.
func ValueOut(val, edgeVal Value) OpenEdge {
	return OpenEdge{fromV: vertex{val: &val}, edgeVal: edgeVal}
}

// TupleOut creates an OpenEdge which, when used to construct a Graph,
// represents an edge (with edgeVal attached to it) coming from the
// TupleVertex comprised of the given ordered-set of input edges.
//
// If len(ins) == 1 && edgeVal.IsZero(), then that single OpenEdge is
// returned as-is.
func TupleOut(ins []OpenEdge, edgeVal Value) OpenEdge {

	if len(ins) == 1 {

		in := ins[0]

		if edgeVal.IsZero() {
			return in
		}

		if in.edgeVal.IsZero() {
			in.edgeVal = edgeVal
			return in
		}

	}

	return OpenEdge{
		fromV:   vertex{tup: ins},
		edgeVal: edgeVal,
	}
}

func (oe OpenEdge) equal(oe2 OpenEdge) bool {
	return oe.edgeVal.Equal(oe2.edgeVal) && oe.fromV.equal(oe2.fromV)
}

type vertex struct {
	val *Value
	tup []OpenEdge
}

func (v vertex) equal(v2 vertex) bool {

	if v.val != nil {
		return v2.val != nil && v.val.Equal(*v2.val)
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

	if v.val != nil {
		return v.val.String()
	}

	strs := make([]string, len(v.tup))

	for i := range v.tup {
		strs[i] = v.tup[i].String()
	}

	return fmt.Sprintf("[%s]", strings.Join(strs, ", "))
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
