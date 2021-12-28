package vm

import "github.com/mediocregopher/ginger/gg"

var (
	inVal  = nameVal("in")
	outVal = nameVal("out")
)

// Operation is an entity which can accept a single argument (the OpenEdge),
// perform some internal processing on that argument, and return a resultant
// Value.
//
// The Scope passed into Perform can be used to Evaluate the OpenEdge, as
// needed.
type Operation interface {
	Perform(gg.OpenEdge, Scope) (Value, error)
}

func preEvalValOp(fn func(Value) (Value, error)) Operation {

	return OperationFunc(func(edge gg.OpenEdge, scope Scope) (Value, error) {

		edgeVal, err := EvaluateEdge(edge, scope)

		if err != nil {
			return Value{}, err
		}

		return fn(edgeVal)
	})
}

// NOTE this is a giant hack to get around the fact that we're not yet
// using a genericized Graph implementation, so when we do AddValueIn
// on a gg.Graph we can't use a Tuple value (because gg has no Tuple
// value), we have to use a Tuple vertex instead.
//
// This also doesn't yet support passing an operation as a value to another
// operation.
func preEvalEdgeOp(fn func(gg.OpenEdge) (Value, error)) Operation {

	return preEvalValOp(func(val Value) (Value, error) {

		var edge gg.OpenEdge

		if len(val.Tuple) > 0 {

			tupEdges := make([]gg.OpenEdge, len(val.Tuple))

			for i := range val.Tuple {
				tupEdges[i] = gg.ValueOut(val.Tuple[i].Value, gg.ZeroValue)
			}

			edge = gg.TupleOut(tupEdges, gg.ZeroValue)

		} else {

			edge = gg.ValueOut(val.Value, gg.ZeroValue)

		}

		return fn(edge)
	})

}

type graphOp struct {
	*gg.Graph
	scope Scope
}

// OperationFromGraph wraps the given Graph such that it can be used as an
// operation.
//
// When Perform is called the passed in OpenEdge is set to the "in" name value
// of the given Graph, then that resultant graph and the given parent Scope are
// used to construct a new Scope. The "out" name value is Evaluated on that
// Scope to obtain a resultant Value.
func OperationFromGraph(g *gg.Graph, scope Scope) Operation {
	return &graphOp{
		Graph: g,
		scope: scope,
	}
}

func (g *graphOp) Perform(edge gg.OpenEdge, scope Scope) (Value, error) {

	return preEvalEdgeOp(func(edge gg.OpenEdge) (Value, error) {

		scope = ScopeFromGraph(
			g.Graph.AddValueIn(edge, inVal.Value),
			g.scope,
		)

		return scope.Evaluate(outVal)

	}).Perform(edge, scope)

}

// OperationFunc is a function which implements the Operation interface.
type OperationFunc func(gg.OpenEdge, Scope) (Value, error)

// Perform calls the underlying OperationFunc directly.
func (f OperationFunc) Perform(edge gg.OpenEdge, scope Scope) (Value, error) {
	return f(edge, scope)
}
