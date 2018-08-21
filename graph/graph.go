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
// Head.
type Edge struct {
	Tail, Head Value
}

func (e Edge) id() string {
	return fmt.Sprintf("%q->%q", e.Tail, e.Head)
}

// an edgeIndex maps valueIDs to a set of edgeIDs. Graph keeps two edgeIndex's,
// one for input edges and one for output edges.
type edgeIndex map[string]map[string]struct{}

func (ei edgeIndex) cp() edgeIndex {
	if ei == nil {
		return edgeIndex{}
	}
	ei2 := make(edgeIndex, len(ei))
	for valID, edgesM := range ei {
		edgesM2 := make(map[string]struct{}, len(edgesM))
		for id := range edgesM {
			edgesM2[id] = struct{}{}
		}
		ei2[valID] = edgesM2
	}
	return ei2
}

func (ei edgeIndex) add(valID, edgeID string) {
	edgesM, ok := ei[valID]
	if !ok {
		edgesM = map[string]struct{}{}
		ei[valID] = edgesM
	}
	edgesM[edgeID] = struct{}{}
}

func (ei edgeIndex) del(valID, edgeID string) {
	edgesM, ok := ei[valID]
	if !ok {
		return
	}

	delete(edgesM, edgeID)
	if len(edgesM) == 0 {
		delete(ei, valID)
	}
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

	// these are indices mapping Value IDs to all the in/out edges for that
	// Value in the Graph.
	vIns, vOuts edgeIndex
}

func (g Graph) cp() Graph {
	g2 := Graph{
		m:     make(map[string]Edge, len(g.m)),
		vIns:  g.vIns.cp(),
		vOuts: g.vOuts.cp(),
	}
	for id, e := range g.m {
		g2.m[id] = e
	}
	return g2
}

// Add returns a new Graph instance with the given Edge added to it. If the
// original Graph already had that Edge this returns the original Graph.
func (g Graph) Add(e Edge) Graph {
	id := e.id()
	if _, ok := g.m[id]; ok {
		return g
	}

	g2 := g.cp()
	g2.m[id] = e
	g2.vIns.add(e.Head.ID, id)
	g2.vOuts.add(e.Tail.ID, id)
	return g2
}

// Del returns a new Graph instance without the given Edge in it. If the
// original Graph didn't have that Edge this returns the original Graph.
func (g Graph) Del(e Edge) Graph {
	id := e.id()
	if _, ok := g.m[id]; !ok {
		return g
	}

	g2 := g.cp()
	delete(g2.m, id)
	g2.vIns.del(e.Head.ID, id)
	g2.vOuts.del(e.Tail.ID, id)
	return g2
}

// Edges returns all Edges which are part of the Graph
func (g Graph) Edges() []Edge {
	edges := make([]Edge, 0, len(g.m))
	for _, e := range g.m {
		edges = append(edges, e)
	}
	return edges
}

// NOTE the Node type exists primarily for convenience. As far as Graph's
// internals are concerned it doesn't _really_ exist, and no Graph method should
// ever take Node as a parameter (except the callback functions like in
// Traverse, where it's not really being taken in).

// Node wraps a Value in a Graph to include that Node's input and output Edges
// in that Graph.
type Node struct {
	Value

	// All Edges in the Graph with this Node's Value as their Head and Tail,
	// respectively.
	Ins, Outs []Edge
}

// Node returns the Node for the given Value, or false if the Graph doesn't
// contain the Value.
func (g Graph) Node(v Value) (Node, bool) {
	n := Node{Value: v}
	for edgeID := range g.vIns[v.ID] {
		n.Ins = append(n.Ins, g.m[edgeID])
	}
	for edgeID := range g.vOuts[v.ID] {
		n.Outs = append(n.Outs, g.m[edgeID])
	}
	return n, len(n.Ins) > 0 || len(n.Outs) > 0
}

// Nodes returns a Node for each Value which has at least one Edge in the Graph,
// with the Nodes mapped by their Value's ID.
func (g Graph) Nodes() map[string]Node {
	nodesM := make(map[string]Node, len(g.m)*2)
	for _, edge := range g.m {
		// if head and tail are modified at the same time it messes up the case
		// where they are the same node
		{
			head := nodesM[edge.Head.ID]
			head.Value = edge.Head
			head.Ins = append(head.Ins, edge)
			nodesM[head.ID] = head
		}
		{
			tail := nodesM[edge.Tail.ID]
			tail.Value = edge.Tail
			tail.Outs = append(tail.Outs, edge)
			nodesM[tail.ID] = tail
		}
	}
	return nodesM
}

// Has returns true if the Graph contains at least one Edge with a Head or Tail
// of Value.
func (g Graph) Has(v Value) bool {
	if _, ok := g.vIns[v.ID]; ok {
		return true
	} else if _, ok := g.vOuts[v.ID]; ok {
		return true
	}
	return false
}

// Traverse is used to traverse the Graph until a stopping point is reached.
// Traversal starts with the cursor at the given start Value. Each hop is
// performed by passing the cursor Value's Node into the next function. The
// cursor moves to the returned Value and next is called again, and so on.
//
// If the boolean returned from the next function is false traversal stops and
// this method returns.
//
// If start has no Edges in the Graph, or a Value returned from next doesn't,
// this will still call next, but the Node will be the zero value.
func (g Graph) Traverse(start Value, next func(n Node) (Value, bool)) {
	curr := start
	for {
		currNode, ok := g.Node(curr)
		if ok {
			curr, ok = next(currNode)
		} else {
			curr, ok = next(Node{})
		}
		if !ok {
			return
		}
	}
}

func (g Graph) edgesShared(g2 Graph) bool {
	for id := range g2.m {
		if _, ok := g.m[id]; !ok {
			return false
		}
	}
	return true
}

// SubGraph returns true if the given Graph shares all of its Edges with this
// Graph.
func (g Graph) SubGraph(g2 Graph) bool {
	// as a quick check before iterating through the edges, if g has fewer edges
	// than g2 then g2 can't possibly be a sub-graph of it
	if len(g.m) < len(g2.m) {
		return false
	}
	return g.edgesShared(g2)
}

// Equal returns true if the given Graph and this Graph have exactly the same
// Edges.
func (g Graph) Equal(g2 Graph) bool {
	if len(g.m) != len(g2.m) {
		return false
	}
	return g.edgesShared(g2)
}
