package vm

import (
	"fmt"

	"github.com/mediocregopher/ginger/gg"
	"github.com/mediocregopher/ginger/graph"
)

// Scope encapsulates a set of names and the values they indicate, or the means
// by which to obtain those values, and allows for the evaluation of a name to
// its value.
type Scope interface {

	// Evaluate accepts a name Value and returns a Thunk which will return the
	// real Value which that name points to.
	Evaluate(Value) (Thunk, error)

	// NewScope returns a new Scope which sub-operations within this Scope
	// should use for themselves.
	NewScope() Scope
}

// EvaluateEdge will use the given Scope to evaluate the edge's ultimate Value,
// after passing all leaf vertices up the tree through all Operations found on
// edge values.
func EvaluateEdge(edge *gg.OpenEdge, scope Scope) (Value, error) {

	thunk, err := graph.MapReduce[gg.Value, gg.Value, Thunk](
		edge,
		func(ggVal gg.Value) (Thunk, error) {

			val := Value{Value: ggVal}

			if val.Name != nil {
				return scope.Evaluate(val)
			}

			return valThunk(val), nil

		},
		func(ggEdgeVal gg.Value, args []Thunk) (Thunk, error) {

			if ggEdgeVal.IsZero() {
				return evalThunks(args), nil
			}

			edgeVal := Value{Value: ggEdgeVal}

			if edgeVal.Name != nil {

				nameThunk, err := scope.Evaluate(edgeVal)

				if err != nil {
					return nil, err

				} else if edgeVal, err = nameThunk(); err != nil {
					return nil, err
				}
			}

			if edgeVal.Graph != nil {

				edgeVal = Value{
					Operation: OperationFromGraph(
						edgeVal.Graph,
						scope.NewScope(),
					),
				}
			}

			if edgeVal.Operation == nil {
				return nil, fmt.Errorf("edge value must be an operation")
			}

			return edgeVal.Operation.Perform(args)
		},
	)

	if err != nil {
		return ZeroValue, err
	}

	return thunk()
}

// ScopeMap implements the Scope interface.
type ScopeMap map[string]Value

var _ Scope = ScopeMap{}

// Evaluate uses the given name Value as a key into the ScopeMap map, and
// returns the Value held there for the key, if any.
func (m ScopeMap) Evaluate(nameVal Value) (Thunk, error) {

	if nameVal.Name == nil {
		return nil, fmt.Errorf("value %v is not a name value", nameVal)
	}

	val, ok := m[*nameVal.Name]

	if !ok {
		return nil, fmt.Errorf("%q not defined", *nameVal.Name)
	}

	return valThunk(val), nil
}

// NewScope returns the ScopeMap as-is.
func (m ScopeMap) NewScope() Scope {
	return m
}

type graphScope struct {
	*gg.Graph
	in Thunk
	parent Scope
}

// ScopeFromGraph returns a Scope which will use the given Graph for evaluation.
//
// When a name is evaluated, that name will be looked up in the Graph. The
// name's vertex must have only a single OpenEdge leading to it. That edge will
// be followed, with edge values being evaluated to Operations, until a Value
// can be obtained.
//
// As a special case, if the name "in" is evaluated, either directly or as part
// of an outer evaluation, then the given Thunk is used to evaluate the Value.
//
// If a name does not appear in the Graph, then the given parent Scope will be
// used to evaluate that name. If the parent Scope is nil then an error is
// returned.
//
// NewScope will return the parent scope, if one is given, or an empty ScopeMap
// if not.
func ScopeFromGraph(g *gg.Graph, in Thunk, parent Scope) Scope {
	return &graphScope{
		Graph:  g,
		in: in,
		parent: parent,
	}
}

func (g *graphScope) Evaluate(nameVal Value) (Thunk, error) {

	if nameVal.Name == nil {
		return nil, fmt.Errorf("value %v is not a name value", nameVal)
	}

	if *nameVal.Name == "in" {
		return g.in, nil
	}

	edgesIn := g.ValueIns(nameVal.Value)

	if l := len(edgesIn); l == 0 && g.parent != nil {

		return g.parent.Evaluate(nameVal)

	} else if l != 1 {

		return nil, fmt.Errorf(
			"%q must have exactly one input edge, found %d input edges",
			*nameVal.Name, l,
		)
	}

	return func() (Value, error) { return EvaluateEdge(edgesIn[0], g) }, nil
}

func (g *graphScope) NewScope() Scope {

	if g.parent == nil {
		return ScopeMap{}
	}

	return g.parent
}
