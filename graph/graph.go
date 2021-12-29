// Package graph implements a generic directed graph type, with support for
// tuple vertices in addition to traditional "value" vertices.
package graph

import (
	"fmt"
	"strings"
)

// Value is any value which can be stored within a Graph. Values should be
// considered immutable, ie once used with the graph package their internal
// value does not change.
type Value interface {
	Equal(Value) bool
	String() string
}

// OpenEdge is an un-realized Edge which can't be used for anything except
// constructing graphs. It has no meaning on its own.
type OpenEdge[V Value] struct {
	val *V
	tup []*OpenEdge[V]

	edgeVal V
}

func (oe *OpenEdge[V]) equal(oe2 *OpenEdge[V]) bool {
	if !oe.edgeVal.Equal(oe2.edgeVal) {
		return false
	}

	if oe.val != nil {
		return oe2.val != nil && (*oe.val).Equal(*oe2.val)
	}

	if len(oe.tup) != len(oe2.tup) {
		return false
	}

	for i := range oe.tup {
		if !oe.tup[i].equal(oe2.tup[i]) {
			return false
		}
	}

	return true
}

func (oe *OpenEdge[V]) String() string {

	vertexType := "tup"

	var fromStr string

	if oe.val != nil {

		vertexType = "val"
		fromStr = (*oe.val).String()

	} else {

		strs := make([]string, len(oe.tup))

		for i := range oe.tup {
			strs[i] = oe.tup[i].String()
		}

		fromStr = fmt.Sprintf("[%s]", strings.Join(strs, ", "))
	}

	return fmt.Sprintf("%s(%s, %s)", vertexType, fromStr, oe.edgeVal.String())
}

// WithEdgeValue returns a copy of the OpenEdge with the given Value replacing
// the previous edge value.
//
// NOTE I _think_ this can be factored out once Graph is genericized.
func (oe *OpenEdge[V]) WithEdgeValue(val V) *OpenEdge[V] {
	oeCp := *oe
	oeCp.edgeVal = val
	return &oeCp
}

// EdgeValue returns the Value which lies on the edge itself.
func (oe OpenEdge[V]) EdgeValue() V {
	return oe.edgeVal
}

// FromValue returns the Value from which the OpenEdge was created via ValueOut,
// or false if it wasn't created via ValueOut.
func (oe OpenEdge[V]) FromValue() (V, bool) {
	if oe.val == nil {
		var zero V
		return zero, false
	}

	return *oe.val, true
}

// FromTuple returns the tuple of OpenEdges from which the OpenEdge was created
// via TupleOut, or false if it wasn't created via TupleOut.
func (oe OpenEdge[V]) FromTuple() ([]*OpenEdge[V], bool) {
	if oe.val != nil {
		return nil, false
	}

	return oe.tup, true
}

// ValueOut creates a OpenEdge which, when used to construct a Graph, represents
// an edge (with edgeVal attached to it) coming from the vertex containing val.
func ValueOut[V Value](val, edgeVal V) *OpenEdge[V] {
	return &OpenEdge[V]{
		val: &val,
		edgeVal: edgeVal,
	}
}

// TupleOut creates an OpenEdge which, when used to construct a Graph,
// represents an edge (with edgeVal attached to it) coming from the vertex
// comprised of the given ordered-set of input edges.
func TupleOut[V Value](ins []*OpenEdge[V], edgeVal V) *OpenEdge[V] {

	if len(ins) == 1 {

		var (
			zero V
			in = ins[0]
		)

		if edgeVal.Equal(zero) {
			return in
		}

		if in.edgeVal.Equal(zero) {
			return in.WithEdgeValue(edgeVal)
		}

	}

	return &OpenEdge[V]{
		tup: ins,
		edgeVal: edgeVal,
	}
}

type graphValueIn[V Value] struct {
	val   V
	edge *OpenEdge[V]
}

func (valIn graphValueIn[V]) equal(valIn2 graphValueIn[V]) bool {
	return valIn.val.Equal(valIn2.val) && valIn.edge.equal(valIn2.edge)
}

// Graph is an immutable container of a set of vertices. The Graph keeps track
// of all Values which terminate an OpenEdge (which may be a tree of Value/Tuple
// vertices).
//
// NOTE The current implementation of Graph is incredibly inefficient, there's
// lots of O(N) operations, unnecessary copying on changes, and duplicate data
// in memory.
type Graph[V Value] struct {
	edges []*OpenEdge[V]
	valIns []graphValueIn[V]
}

func (g *Graph[V]) cp() *Graph[V] {
	cp := &Graph[V]{
		edges: make([]*OpenEdge[V], len(g.edges)),
		valIns: make([]graphValueIn[V], len(g.valIns)),
	}
	copy(cp.edges, g.edges)
	copy(cp.valIns, g.valIns)
	return cp
}

func (g *Graph[V]) String() string {

	var strs []string

	for _, valIn := range g.valIns {
		strs = append(
			strs,
			fmt.Sprintf("valIn(%s, %s)", valIn.edge.String(), valIn.val.String()),
		)
	}

	return fmt.Sprintf("graph(%s)", strings.Join(strs, ", "))
}

// NOTE this method is used more for its functionality than for any performance
// reasons... it's incredibly inefficient in how it deduplicates edges, but by
// doing the deduplication we enable the graph map operation to work correctly.
func (g *Graph[V]) dedupeEdge(edge *OpenEdge[V]) *OpenEdge[V] {

	// check if there's an existing edge which is fully equivalent in the graph
	// already, and if so return that.
	for i := range g.edges {
		if g.edges[i].equal(edge) {
			return g.edges[i]
		}
	}

	// if this edge is a value edge then there's nothing else to do, return it.
	if _, ok := edge.FromValue(); ok {
		return edge
	}

	// this edge is a tuple edge, it's possible that one of its sub-edges is
	// already in the graph. dedupe each sub-edge individually.

	tupEdges := make([]*OpenEdge[V], len(edge.tup))

	for i := range edge.tup {
		tupEdges[i] = g.dedupeEdge(edge.tup[i])
	}

	return TupleOut(tupEdges, edge.EdgeValue())
}

// ValueIns returns, if any, all OpenEdges which lead to the given Value in the
// Graph (ie, all those added via AddValueIn).
//
// The returned slice should not be modified.
func (g *Graph[V]) ValueIns(val Value) []*OpenEdge[V] {

	var edges []*OpenEdge[V]

	for _, valIn := range g.valIns {
		if valIn.val.Equal(val) {
			edges = append(edges, valIn.edge)
		}
	}

	return edges
}

// AddValueIn takes a OpenEdge and connects it to the Value vertex containing
// val, returning the new Graph which reflects that connection.
func (g *Graph[V]) AddValueIn(oe *OpenEdge[V], val V) *Graph[V] {

	valIn := graphValueIn[V]{
		val: val,
		edge: oe,
	}

	for i := range g.valIns {
		if g.valIns[i].equal(valIn) {
			return g
		}
	}

	valIn.edge = g.dedupeEdge(valIn.edge)

	g = g.cp()
	g.valIns = append(g.valIns, valIn)

	return g
}

// Equal returns whether or not the two Graphs are equivalent in value.
func (g *Graph[V]) Equal(g2 *Graph[V]) bool {

	if len(g.valIns) != len(g2.valIns) {
		return false
	}

outer:
	for _, valIn := range g.valIns {

		for _, valIn2 := range g2.valIns {

			if valIn.equal(valIn2) {
				continue outer
			}
		}

		return false
	}

	return true
}
