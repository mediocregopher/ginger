// Package constraint implements an extremely simple constraint engine.
// Elements, and constraints on those elements, are given to the engine, which
// uses those constraints to generate an output. Elements are defined as a
// string
package constraint

import (
	"github.com/mediocregopher/ginger/gg"
)

// Constraint describes a constraint on an element. The Elem field must be
// filled in, as well as exactly one other field
type Constraint struct {
	Elem string

	// LT says that Elem is less than this element
	LT string
}

const ltEdge = gg.Str("lt")

// Engine processes sets of constraints to generate an output
type Engine struct {
	g *gg.Graph
}

// NewEngine initializes and returns an empty Engine
func NewEngine() *Engine {
	return &Engine{g: gg.Null}
}

// AddConstraint adds the given constraint to the engine's set, returns false if
// the constraint couldn't be added due to a conflict with a previous constraint
func (e *Engine) AddConstraint(c Constraint) bool {
	elem := gg.Str(c.Elem)
	g := e.g.AddValueIn(gg.ValueOut(elem, ltEdge), gg.Str(c.LT))

	// Check for loops in g starting at c.Elem, bail if there are any
	{
		seen := map[*gg.Vertex]bool{}
		start := g.Value(elem)
		var hasLoop func(v *gg.Vertex) bool
		hasLoop = func(v *gg.Vertex) bool {
			if seen[v] {
				return v == start
			}
			seen[v] = true
			for _, out := range v.Out {
				if hasLoop(out.To) {
					return true
				}
			}
			return false
		}
		if hasLoop(start) {
			return false
		}
	}

	e.g = g
	return true
}

// Solve uses the constraints which have been added to the engine to give a
// possible solution. The given element is one which has been added to the
// engine and whose value is known to be zero.
func (e *Engine) Solve() map[string]int {
	m := map[string]int{}
	if len(e.g.Values()) == 0 {
		return m
	}

	vElem := func(v *gg.Vertex) string {
		return string(v.Value.(gg.Str))
	}

	// first the roots are determined to be the elements with no In edges, which
	// _must_ exist since the graph presumably has no loops
	var roots []*gg.Vertex
	e.g.Walk(nil, func(v *gg.Vertex) bool {
		if len(v.In) == 0 {
			roots = append(roots, v)
			m[vElem(v)] = 0
		}
		return true
	})

	// sanity check
	if len(roots) == 0 {
		panic("no roots found in graph somehow")
	}

	// a vertex's value is then the length of the longest path from it to one of
	// the roots
	var walk func(*gg.Vertex, int)
	walk = func(v *gg.Vertex, val int) {
		if elem := vElem(v); val > m[elem] {
			m[elem] = val
		}
		for _, out := range v.Out {
			walk(out.To, val+1)
		}
	}
	for _, root := range roots {
		walk(root, 0)
	}

	return m
}
