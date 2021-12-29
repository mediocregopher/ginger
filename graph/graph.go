// Package graph implements a generic directed graph type, with support for
// tuple vertices in addition to traditional "value" vertices.
package graph

import (
	"fmt"
	"strings"
)

// Value is any value which can be stored within a Graph.
type Value interface {
	Equal(Value) bool
	String() string
}

// OpenEdge is an un-realized Edge which can't be used for anything except
// constructing graphs. It has no meaning on its own.
type OpenEdge[V Value] struct {
	fromV   vertex[V]
	edgeVal V
}

func (oe OpenEdge[V]) equal(oe2 OpenEdge[V]) bool {
	return oe.edgeVal.Equal(oe2.edgeVal) && oe.fromV.equal(oe2.fromV)
}

func (oe OpenEdge[V]) String() string {

	vertexType := "tup"

	if oe.fromV.val != nil {
		vertexType = "val"
	}

	return fmt.Sprintf("%s(%s, %s)", vertexType, oe.fromV.String(), oe.edgeVal.String())
}

// WithEdgeValue returns a copy of the OpenEdge with the given Value replacing
// the previous edge value.
//
// NOTE I _think_ this can be factored out once Graph is genericized.
func (oe OpenEdge[V]) WithEdgeValue(val V) OpenEdge[V] {
	oe.edgeVal = val
	return oe
}

// EdgeValue returns the Value which lies on the edge itself.
func (oe OpenEdge[V]) EdgeValue() V {
	return oe.edgeVal
}

// FromValue returns the Value from which the OpenEdge was created via ValueOut,
// or false if it wasn't created via ValueOut.
func (oe OpenEdge[V]) FromValue() (V, bool) {
	if oe.fromV.val == nil {
		var zero V
		return zero, false
	}

	return *oe.fromV.val, true
}

// FromTuple returns the tuple of OpenEdges from which the OpenEdge was created
// via TupleOut, or false if it wasn't created via TupleOut.
func (oe OpenEdge[V]) FromTuple() ([]OpenEdge[V], bool) {
	if oe.fromV.val != nil {
		return nil, false
	}

	return oe.fromV.tup, true
}

// ValueOut creates a OpenEdge which, when used to construct a Graph, represents
// an edge (with edgeVal attached to it) coming from the ValueVertex containing
// val.
func ValueOut[V Value](val, edgeVal V) OpenEdge[V] {
	return OpenEdge[V]{fromV: vertex[V]{val: &val}, edgeVal: edgeVal}
}

// TupleOut creates an OpenEdge which, when used to construct a Graph,
// represents an edge (with edgeVal attached to it) coming from the
// TupleVertex comprised of the given ordered-set of input edges.
//
// If len(ins) == 1 && edgeVal.IsZero(), then that single OpenEdge is
// returned as-is.
func TupleOut[V Value](ins []OpenEdge[V], edgeVal V) OpenEdge[V] {

	if len(ins) == 1 {

		in := ins[0]
		var zero V

		if edgeVal.Equal(zero) {
			return in
		}

		if in.edgeVal.Equal(zero) {
			in.edgeVal = edgeVal
			return in
		}

	}

	return OpenEdge[V]{
		fromV:   vertex[V]{tup: ins},
		edgeVal: edgeVal,
	}
}


type vertex[V Value] struct {
	val *V
	tup []OpenEdge[V]
}

func (v vertex[V]) equal(v2 vertex[V]) bool {

	if v.val != nil {
		return v2.val != nil && (*v.val).Equal(*v2.val)
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

func (v vertex[V]) String() string {

	if v.val != nil {
		return (*v.val).String()
	}

	strs := make([]string, len(v.tup))

	for i := range v.tup {
		strs[i] = v.tup[i].String()
	}

	return fmt.Sprintf("[%s]", strings.Join(strs, ", "))
}

type graphValueIn[V Value] struct {
	val   V
	edges []OpenEdge[V]
}

func (valIn graphValueIn[V]) cp() graphValueIn[V] {
	cp := valIn
	cp.edges = make([]OpenEdge[V], len(valIn.edges))
	copy(cp.edges, valIn.edges)
	return valIn
}

func (valIn graphValueIn[V]) equal(valIn2 graphValueIn[V]) bool {
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
type Graph[V Value] struct {
	valIns []graphValueIn[V]
}

func (g *Graph[V]) cp() *Graph[V] {
	cp := &Graph[V]{
		valIns: make([]graphValueIn[V], len(g.valIns)),
	}
	copy(cp.valIns, g.valIns)
	return cp
}

func (g *Graph[V]) String() string {

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

// ValueIns returns, if any, all OpenEdges which lead to the given Value in the
// Graph (ie, all those added via AddValueIn).
func (g *Graph[V]) ValueIns(val Value) []OpenEdge[V] {
	for _, valIn := range g.valIns {
		if valIn.val.Equal(val) {
			return valIn.cp().edges
		}
	}

	return nil
}

// AddValueIn takes a OpenEdge and connects it to the Value vertex containing
// val, returning the new Graph which reflects that connection.
func (g *Graph[V]) AddValueIn(oe OpenEdge[V], val V) *Graph[V] {

	edges := g.ValueIns(val)

	for _, existingOE := range edges {
		if existingOE.equal(oe) {
			return g
		}
	}

	// ValueIns returns a copy of edges, so we're ok to modify it.
	edges = append(edges, oe)
	valIn := graphValueIn[V]{val: val, edges: edges}

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
