package vm

import (
	"fmt"

	"github.com/mediocregopher/ginger/gg"
)

// Scope encapsulates a set of name->Value mappings.
type Scope interface {

	// Resolve accepts a name and returns an Value.
	Resolve(string) (Value, error)

	// NewScope returns a new Scope which sub-operations within this Scope
	// should use for themselves.
	NewScope() Scope
}

// ScopeMap implements the Scope interface.
type ScopeMap map[string]Value

var _ Scope = ScopeMap{}

// Resolve uses the given name as a key into the ScopeMap map, and
// returns the Operation held there for the key, if any.
func (m ScopeMap) Resolve(name string) (Value, error) {

	v, ok := m[name]

	if !ok {
		return Value{}, fmt.Errorf("%q not defined", name)
	}

	return v, nil
}

// NewScope returns the ScopeMap as-is.
func (m ScopeMap) NewScope() Scope {
	return m
}

type scopeWith struct {
	Scope // parent
	name  string
	val   Value
}

// ScopeWith returns a copy of the given Scope, except that evaluating the given
// name will always return the given Value.
func ScopeWith(scope Scope, name string, val Value) Scope {
	return &scopeWith{
		Scope: scope,
		name:  name,
		val:   val,
	}
}

func (s *scopeWith) Resolve(name string) (Value, error) {
	if name == s.name {
		return s.val, nil
	}
	return s.Scope.Resolve(name)
}

type graphScope struct {
	*gg.Graph
	parent Scope
}

/*

TODO I don't think this is actually necessary

// ScopeFromGraph returns a Scope which will use the given Graph for name
// resolution.
//
// When a name is resolved, that name will be looked up in the Graph. The name's
// vertex must have only a single OpenEdge leading to it. That edge will be
// compiled into an Operation and returned.
//
// If a name does not appear in the Graph, then the given parent Scope will be
// used to resolve that name. If the parent Scope is nil then an error is
// returned.
//
// NewScope will return the parent scope, if one is given, or an empty ScopeMap
// if not.
func ScopeFromGraph(g *gg.Graph, parent Scope) Scope {
	return &graphScope{
		Graph:  g,
		parent: parent,
	}
}

func (g *graphScope) Resolve(name string) (Value, error) {

	var ggNameVal gg.Value
	ggNameVal.Name = &name

	log.Printf("resolving %q", name)
	edgesIn := g.ValueIns(ggNameVal)

	if l := len(edgesIn); l == 0 && g.parent != nil {

		return g.parent.Resolve(name)

	} else if l != 1 {

		return nil, fmt.Errorf(
			"%q must have exactly one input edge, found %d input edges",
			name, l,
		)
	}

	return CompileEdge(edgesIn[0], g)
}

func (g *graphScope) NewScope() Scope {

	if g.parent == nil {
		return ScopeMap{}
	}

	return g.parent
}

*/
