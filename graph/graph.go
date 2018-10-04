// Package graph implements an immutable unidirectional graph.
package graph

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
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
type Edge interface {
	Tail() Value // The Value the Edge is coming from
	Head() Value // The Value the Edge is going to
}

func edgeID(e Edge) string {
	return fmt.Sprintf("%q->%q", e.Tail().ID, e.Head().ID)
}

type edge struct {
	tail, head Value
}

// NewEdge constructs and returns an Edge running from tail to head.
func NewEdge(tail, head Value) Edge {
	return edge{tail, head}
}

func (e edge) Tail() Value {
	return e.tail
}

func (e edge) Head() Value {
	return e.head
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

func (g Graph) String() string {
	edgeIDs := make([]string, 0, len(g.m))
	for edgeID := range g.m {
		edgeIDs = append(edgeIDs, edgeID)
	}
	sort.Strings(edgeIDs)
	return "Graph{" + strings.Join(edgeIDs, ",") + "}"
}

// Add returns a new Graph instance with the given Edge added to it. If the
// original Graph already had that Edge this returns the original Graph.
func (g Graph) Add(e Edge) Graph {
	id := edgeID(e)
	if _, ok := g.m[id]; ok {
		return g
	}

	g2 := g.cp()
	g2.addDirty(id, e)
	return g2
}

func (g Graph) addDirty(edgeID string, e Edge) {
	g.m[edgeID] = e
	g.vIns.add(e.Head().ID, edgeID)
	g.vOuts.add(e.Tail().ID, edgeID)
}

func (g Graph) estSize() int {
	lvIns := len(g.vIns)
	lvOuts := len(g.vOuts)
	if lvIns > lvOuts {
		return lvIns
	}
	return lvOuts
}

// Del returns a new Graph instance without the given Edge in it. If the
// original Graph didn't have that Edge this returns the original Graph.
func (g Graph) Del(e Edge) Graph {
	id := edgeID(e)
	if _, ok := g.m[id]; !ok {
		return g
	}

	g2 := g.cp()
	delete(g2.m, id)
	g2.vIns.del(e.Head().ID, id)
	g2.vOuts.del(e.Tail().ID, id)
	return g2
}

// Disjoin looks at the whole Graph and returns all sub-graphs of it which don't
// share any Edges between each other.
func (g Graph) Disjoin() []Graph {
	valM := make(map[string]*Graph, len(g.vOuts))
	graphForEdge := func(edge Edge) *Graph {
		headGraph := valM[edge.Head().ID]
		tailGraph := valM[edge.Tail().ID]
		if headGraph == nil && tailGraph == nil {
			newGraph := Graph{}.cp() // cp also initializes
			return &newGraph
		} else if headGraph == nil && tailGraph != nil {
			return tailGraph
		} else if headGraph != nil && tailGraph == nil {
			return headGraph
		} else if headGraph == tailGraph {
			return headGraph // doesn't matter which is returned
		}

		// the two values are part of different graphs, join the smaller into
		// the larger and change all values which were pointing to it to point
		// into the larger (which will then be the join of them)
		if len(tailGraph.m) > len(headGraph.m) {
			tailGraph, headGraph = headGraph, tailGraph
		}
		for edgeID, edge := range tailGraph.m {
			(*headGraph).addDirty(edgeID, edge)
		}
		for valID, valGraph := range valM {
			if valGraph == tailGraph {
				valM[valID] = headGraph
			}
		}
		return headGraph
	}

	for edgeID, edge := range g.m {
		graph := graphForEdge(edge)
		(*graph).addDirty(edgeID, edge)
		valM[edge.Head().ID] = graph
		valM[edge.Tail().ID] = graph
	}

	found := map[*Graph]bool{}
	graphs := make([]Graph, 0, len(valM))
	for _, graph := range valM {
		if found[graph] {
			continue
		}
		found[graph] = true
		graphs = append(graphs, *graph)
	}
	return graphs
}

// Join returns a new Graph which shares all Edges of this Graph and all given
// Graphs.
func (g Graph) Join(graphs ...Graph) Graph {
	g2 := g.cp()
	for _, graph := range graphs {
		for edgeID, edge := range graph.m {
			g2.addDirty(edgeID, edge)
		}
	}
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
	// respectively. These should not be expected to be deterministic.
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
			headV := edge.Head()
			head := nodesM[headV.ID]
			head.Value = headV
			head.Ins = append(head.Ins, edge)
			nodesM[head.ID] = head
		}
		{
			tailV := edge.Tail()
			tail := nodesM[tailV.ID]
			tail.Value = tailV
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

// VisitBreadth is like Traverse, except that each Node is only visited once,
// and the order of visited Nodes is determined by traversing each Node's output
// Edges breadth-wise.
//
// If the boolean returned from the callback function is false, or the start
// Value has no edges in the Graph, traversal stops and this method returns.
//
// The exact order of Nodes visited is _not_ deterministic.
func (g Graph) VisitBreadth(start Value, callback func(n Node) bool) {
	visited := map[string]bool{}
	toVisit := make([]Value, 0, g.estSize())
	toVisit = append(toVisit, start)

	for {
		if len(toVisit) == 0 {
			return
		}

		// shift val off front
		val := toVisit[0]
		toVisit = toVisit[1:]
		if visited[val.ID] {
			continue
		}
		node, ok := g.Node(val)
		if !ok {
			continue
		} else if !callback(node) {
			return
		}
		visited[val.ID] = true
		for _, edge := range node.Outs {
			headV := edge.Head()
			if visited[headV.ID] {
				continue
			}
			toVisit = append(toVisit, headV)
		}
	}
}

// VisitDepth is like Traverse, except that each Node is only visited once,
// and the order of visited Nodes is determined by traversing each Node's output
// Edges depth-wise.
//
// If the boolean returned from the callback function is false, or the start
// Value has no edges in the Graph, traversal stops and this method returns.
//
// The exact order of Nodes visited is _not_ deterministic.
func (g Graph) VisitDepth(start Value, callback func(n Node) bool) {
	// VisitDepth is actually the same as VisitBreadth, only you read off the
	// toVisit list from back-to-front
	visited := map[string]bool{}
	toVisit := make([]Value, 0, g.estSize())
	toVisit = append(toVisit, start)

	for {
		if len(toVisit) == 0 {
			return
		}

		val := toVisit[0]
		toVisit = toVisit[:len(toVisit)-1] // pop val off back
		if visited[val.ID] {
			continue
		}
		node, ok := g.Node(val)
		if !ok {
			continue
		} else if !callback(node) {
			return
		}
		visited[val.ID] = true
		for _, edge := range node.Outs {
			if visited[edge.Head().ID] {
				continue
			}
			toVisit = append(toVisit, edge.Head())
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
