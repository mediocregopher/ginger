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

// OpenEdge consists of the edge value (E) and source vertex value (V) of an
// edge in a Graph. When passed into the AddValueIn method a full edge is
// created. An OpenEdge can also be sourced from a tuple vertex, whose value is
// an ordered set of OpenEdges of this same type.
type OpenEdge[E, V Value] struct {
	val *V
	tup []*OpenEdge[E, V]

	edgeVal E
}

func (oe *OpenEdge[E, V]) equal(oe2 *OpenEdge[E, V]) bool {
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

func (oe *OpenEdge[E, V]) String() string {

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
func (oe *OpenEdge[E, V]) WithEdgeValue(val E) *OpenEdge[E, V] {
	oeCp := *oe
	oeCp.edgeVal = val
	return &oeCp
}

// EdgeValue returns the Value which lies on the edge itself.
func (oe OpenEdge[E, V]) EdgeValue() E {
	return oe.edgeVal
}

// FromValue returns the Value from which the OpenEdge was created via ValueOut,
// or false if it wasn't created via ValueOut.
func (oe OpenEdge[E, V]) FromValue() (V, bool) {
	if oe.val == nil {
		var zero V
		return zero, false
	}

	return *oe.val, true
}

// FromTuple returns the tuple of OpenEdges from which the OpenEdge was created
// via TupleOut, or false if it wasn't created via TupleOut.
func (oe OpenEdge[E, V]) FromTuple() ([]*OpenEdge[E, V], bool) {
	if oe.val != nil {
		return nil, false
	}

	return oe.tup, true
}

// ValueOut creates a OpenEdge which, when used to construct a Graph, represents
// an edge (with edgeVal attached to it) coming from the vertex containing val.
func ValueOut[E, V Value](edgeVal E, val V) *OpenEdge[E, V] {
	return &OpenEdge[E, V]{
		val: &val,
		edgeVal: edgeVal,
	}
}

// TupleOut creates an OpenEdge which, when used to construct a Graph,
// represents an edge (with edgeVal attached to it) coming from the vertex
// comprised of the given ordered-set of input edges.
func TupleOut[E, V Value](edgeVal E, ins ...*OpenEdge[E, V]) *OpenEdge[E, V] {

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

	return &OpenEdge[E, V]{
		tup: ins,
		edgeVal: edgeVal,
	}
}

type graphValueIn[E, V Value] struct {
	val   V
	edge *OpenEdge[E, V]
}

func (valIn graphValueIn[E, V]) equal(valIn2 graphValueIn[E, V]) bool {
	return valIn.val.Equal(valIn2.val) && valIn.edge.equal(valIn2.edge)
}

// Graph is an immutable container of a set of vertices. The Graph keeps track
// of all Values which terminate an OpenEdge. E indicates the type of edge
// values, while V indicates the type of vertex values.
//
// NOTE The current implementation of Graph is incredibly inefficient, there's
// lots of O(N) operations, unnecessary copying on changes, and duplicate data
// in memory.
type Graph[E, V Value] struct {
	edges []*OpenEdge[E, V]
	valIns []graphValueIn[E, V]
}

func (g *Graph[E, V]) cp() *Graph[E, V] {
	cp := &Graph[E, V]{
		edges: make([]*OpenEdge[E, V], len(g.edges)),
		valIns: make([]graphValueIn[E, V], len(g.valIns)),
	}
	copy(cp.edges, g.edges)
	copy(cp.valIns, g.valIns)
	return cp
}

func (g *Graph[E, V]) String() string {

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
func (g *Graph[E, V]) dedupeEdge(edge *OpenEdge[E, V]) *OpenEdge[E, V] {

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

	tupEdges := make([]*OpenEdge[E, V], len(edge.tup))

	for i := range edge.tup {
		tupEdges[i] = g.dedupeEdge(edge.tup[i])
	}

	return TupleOut(edge.EdgeValue(), tupEdges...)
}

// ValueIns returns, if any, all OpenEdges which lead to the given Value in the
// Graph (ie, all those added via AddValueIn).
//
// The returned slice should not be modified.
func (g *Graph[E, V]) ValueIns(val Value) []*OpenEdge[E, V] {

	var edges []*OpenEdge[E, V]

	for _, valIn := range g.valIns {
		if valIn.val.Equal(val) {
			edges = append(edges, valIn.edge)
		}
	}

	return edges
}

// AddValueIn takes a OpenEdge and connects it to the Value vertex containing
// val, returning the new Graph which reflects that connection.
func (g *Graph[E, V]) AddValueIn(val V, oe *OpenEdge[E, V]) *Graph[E, V] {

	valIn := graphValueIn[E, V]{
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
func (g *Graph[E, V]) Equal(g2 *Graph[E, V]) bool {

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


func mapReduce[Ea, Va Value, Vb any](
	root *OpenEdge[Ea, Va],
	mapVal func(Va) (Vb, error),
	reduceEdge func(*OpenEdge[Ea, Va], []Vb) (Vb, error),
) (
	Vb, error,
){

	if valA, ok := root.FromValue(); ok {

		valB, err := mapVal(valA)

		if err != nil {
			var zero Vb
			return zero, err
		}

		return reduceEdge(root, []Vb{valB})
	}

	tupA, _ := root.FromTuple()

	valsB := make([]Vb, len(tupA))

	for i := range tupA {

		valB, err := mapReduce[Ea, Va, Vb](
			tupA[i], mapVal, reduceEdge,
		)

		if err != nil {
			var zero Vb
			return zero, err
		}

		valsB[i] = valB
	}

	return reduceEdge(root, valsB)
}

type mappedVal[Va Value, Vb any] struct {
	valA Va
	valB Vb // result
}

type reducedEdge[Ea, Va Value, Vb any] struct {
	edgeA *OpenEdge[Ea, Va]
	valB Vb // result
}

// MapReduce recursively computes a resultant Value of type Vb from an
// OpenEdge[Ea, Va].
//
// Tuple edges which are encountered will have Reduce called on each OpenEdge
// branch of the tuple, to obtain a Vb for each branch. The edge value of the
// tuple edge (Ea) and the just obtained Vbs are then passed to reduceEdge to
// obtain a Vb for that edge.
//
// The values of value edges (Va) which are encountered are mapped to Vb using
// the mapVal function. The edge value of those value edges (Ea) and the just
// obtained Vb value are then passed to reduceEdge to obtain a Vb for that edge.
//
// If either the map or reduce function returns an error then processing is
// immediately cancelled and that error is returned directly.
//
// If a value or edge is connected to multiple times within the root OpenEdge it
// will only be mapped/reduced a single time, and the result of that single
// map/reduction will be passed to each dependant operation.
//
func MapReduce[Ea, Va Value, Vb any](
	root *OpenEdge[Ea, Va],
	mapVal func(Va) (Vb, error),
	reduceEdge func(Ea, []Vb) (Vb, error),
) (
	Vb, error,
){

	var (
		zeroB Vb

		// we use these to memoize reductions on values and edges, so a
		// reduction is only performed a single time for each value/edge.
		//
		// NOTE this is not implemented very efficiently.
		mappedVals []mappedVal[Va, Vb]
		reducedEdges []reducedEdge[Ea, Va, Vb]
	)

	return mapReduce[Ea, Va, Vb](
		root,
		func(valA Va) (Vb, error) {

			for _, mappedVal := range mappedVals {
				if mappedVal.valA.Equal(valA) {
					return mappedVal.valB, nil
				}
			}

			valB, err := mapVal(valA)

			if err != nil {
				return zeroB, err
			}

			mappedVals = append(mappedVals, mappedVal[Va, Vb]{
				valA: valA,
				valB: valB,
			})

			return valB, nil
		},
		func(edgeA *OpenEdge[Ea, Va], valBs []Vb) (Vb, error) {

			for _, reducedEdge := range reducedEdges {
				if reducedEdge.edgeA.equal(edgeA) {
					return reducedEdge.valB, nil
				}
			}

			valB, err := reduceEdge(edgeA.EdgeValue(), valBs)

			if err != nil {
				return zeroB, err
			}

			reducedEdges = append(reducedEdges, reducedEdge[Ea, Va, Vb]{
				edgeA: edgeA,
				valB: valB,
			})

			return valB, nil
		},
	)
}
