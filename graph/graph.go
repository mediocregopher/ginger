// Package graph implements an immutable unidirectional graph.
package graph

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// Value wraps a go value in a way such that it will be uniquely identified
// within any Graph and between Graphs. Use NewValue to create a Value instance.
// You can create an instance manually as long as ID is globally unique.
type Value struct {
	ID string
	V  interface{}
}

// Void is the absence of any value.
var Void Value

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

// Edge is a directional edge connecting two values in a Graph, the Tail and the
// Head. An Edge may also contain a value of its own.
type Edge struct {
	Tail, Val, Head Value
}

func (e Edge) id() string {
	return fmt.Sprintf("%q-%q->%q", e.Tail, e.Val, e.Head)
}

// Graph implements an immutable, unidirectional graph which can hold generic
// values. All methods are thread-safe as they don't modify the Graph in any
// way.
//
// The Graph's zero value is the initial empty graph.
//
// The Graph does not keep track of Edge ordering. Assume that all slices of
// Edges are in random order.
type Graph struct {
	m map[string]Edge
}

func (g Graph) cp() Graph {
	g2 := Graph{
		m: make(map[string]Edge, len(g.m)),
	}
	for id, e := range g.m {
		g2.m[id] = e
	}
	return g2
}

// AddEdge returns a new Graph instance with the given Edge added to it. If the
// original Graph already had that Edge this returns the original Graph.
func (g Graph) AddEdge(e Edge) Graph {
	id := e.id()
	if _, ok := g.m[id]; ok {
		return g
	}

	g2 := g.cp()
	g2.m[id] = e
	return g2
}

// DelEdge returns a new Graph instance without the given Edge in it. If the
// original Graph didn't have that Edge  this returns the original Graph.
func (g Graph) DelEdge(e Edge) Graph {
	id := e.id()
	if _, ok := g.m[id]; !ok {
		return g
	}

	g2 := g.cp()
	delete(g2.m, id)
	return g2
}

// Values returns all Values which have incoming or outgoing Edges in the Graph.
func (g Graph) Values() []Value {
	values := make([]Value, 0, len(g.m))
	found := map[string]bool{}
	tryAdd := func(v Value) {
		if ok := found[v.ID]; !ok {
			values = append(values, v)
			found[v.ID] = true
		}
	}

	for _, e := range g.m {
		tryAdd(e.Head)
		tryAdd(e.Tail)
	}
	return values
}

// Edges returns all Edges which are part of the Graph
func (g Graph) Edges() []Edge {
	edges := make([]Edge, 0, len(g.m))
	for _, e := range g.m {
		edges = append(edges, e)
	}
	return edges
}

// ValueEdges returns all input (e.Head==v) and output (e.Tail==v) Edges
// for the given Value in the Graph.
func (g Graph) ValueEdges(v Value) ([]Edge, []Edge) {
	var in, out []Edge
	for _, e := range g.m {
		if e.Tail.ID == v.ID {
			out = append(out, e)
		}
		if e.Head.ID == v.ID {
			in = append(in, e)
		}
	}
	return in, out
}

// Traverse is used to traverse the Graph until a stopping point is reached.
// Traversal starts with the cursor at the given start value. Each hop is
// performed by passing the cursor value along with its input and output Edges
// into the next function. The cursor moves to the returned Value and next is
// called again, and so on.
//
// If the boolean returned from the next function is false traversal stops and
// this method returns.
//
// If start has no Edges in the Graph, or a Value returned from next doesn't,
// this will still call next, but the in/out params will both be empty.
func (g Graph) Traverse(start Value, next func(v Value, in, out []Edge) (Value, bool)) {
	curr := start
	var ok bool
	for {
		in, out := g.ValueEdges(curr)
		if curr, ok = next(curr, in, out); !ok {
			return
		}
	}
}