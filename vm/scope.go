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

	// Evaluate accepts a name Value and returns the real Value which that name
	// points to.
	Evaluate(Value) (Value, error)

	// NewScope returns a new Scope which sub-operations within this Scope
	// should use for themselves.
	NewScope() Scope
}

// edgeToValue ignores the edgeValue, it only evaluates the edge's vertex as a
// Value.
func edgeToValue(edge graph.OpenEdge[gg.Value], scope Scope) (Value, error) {

	if ggVal, ok := edge.FromValue(); ok {

		val := Value{Value: ggVal}

		if val.Name != nil {
			return scope.Evaluate(val)
		}

		return val, nil
	}

	var tupVal Value

	tup, _ := edge.FromTuple()

	for _, tupEdge := range tup {

		val, err := EvaluateEdge(tupEdge, scope)

		if err != nil {
			return Value{}, err
		}

		tupVal.Tuple = append(tupVal.Tuple, val)
	}

	if len(tupVal.Tuple) == 1 {
		return tupVal.Tuple[0], nil
	}

	return tupVal, nil
}

// EvaluateEdge will use the given Scope to evaluate the edge's ultimate Value,
// after passing all leaf vertices up the tree through all Operations found on
// edge values.
func EvaluateEdge(edge graph.OpenEdge[gg.Value], scope Scope) (Value, error) {

	edgeVal := Value{Value: edge.EdgeValue()}

	if edgeVal.IsZero() {
		return edgeToValue(edge, scope)
	}

	edge = edge.WithEdgeValue(gg.ZeroValue)

	if edgeVal.Name != nil {

		var err error

		if edgeVal, err = scope.Evaluate(edgeVal); err != nil {
			return Value{}, err
		}
	}

	if edgeVal.Graph != nil {

		edgeVal = Value{
			Operation: OperationFromGraph(edgeVal.Graph, scope.NewScope()),
		}
	}

	if edgeVal.Operation == nil {
		return Value{}, fmt.Errorf("edge value must be an operation")
	}

	return edgeVal.Operation.Perform(edge, scope)
}

// ScopeMap implements the Scope interface.
type ScopeMap map[string]Value

var _ Scope = ScopeMap{}

// Evaluate uses the given name Value as a key into the ScopeMap map, and
// returns the Value held there for the key, if any.
func (m ScopeMap) Evaluate(nameVal Value) (Value, error) {

	if nameVal.Name == nil {
		return Value{}, fmt.Errorf("value %v is not a name value", nameVal)
	}

	val, ok := m[*nameVal.Name]

	if !ok {
		return Value{}, fmt.Errorf("%q not defined", *nameVal.Name)
	}

	return val, nil
}

// NewScope returns the ScopeMap as-is.
func (m ScopeMap) NewScope() Scope {
	return m
}

type graphScope struct {
	*graph.Graph[gg.Value]
	parent Scope
}

// ScopeFromGraph returns a Scope which will use the given Graph for evaluation.
//
// When a name is evaluated, that name will be looked up in the Graph. The
// name's vertex must have only a single OpenEdge leading to it. That edge will
// be followed, with edge values being evaluated to Operations, until a Value
// can be obtained.
//
// If a name does not appear in the Graph, then the given parent Scope will be
// used to evaluate that name. If the parent Scope is nil then an error is
// returned.
//
// NewScope will return the parent scope, if one is given, or an empty ScopeMap
// if not.
func ScopeFromGraph(g *graph.Graph[gg.Value], parent Scope) Scope {
	return &graphScope{
		Graph:  g,
		parent: parent,
	}
}

func (g *graphScope) Evaluate(nameVal Value) (Value, error) {

	if nameVal.Name == nil {
		return Value{}, fmt.Errorf("value %v is not a name value", nameVal)
	}

	edgesIn := g.ValueIns(nameVal.Value)

	if l := len(edgesIn); l == 0 && g.parent != nil {

		return g.parent.Evaluate(nameVal)

	} else if l != 1 {

		return Value{}, fmt.Errorf(
			"%q must have exactly one input edge, found %d input edges",
			*nameVal.Name, l,
		)
	}

	return EvaluateEdge(edgesIn[0], g)
}

func (g *graphScope) NewScope() Scope {

	if g.parent == nil {
		return ScopeMap{}
	}

	return g.parent
}
