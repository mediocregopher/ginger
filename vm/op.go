package vm

import (
	"github.com/mediocregopher/ginger/gg"
	"github.com/mediocregopher/ginger/graph"
)

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
	Perform(*graph.OpenEdge[gg.Value], Scope) (Value, error)
}

func preEvalValOp(fn func(Value) (Value, error)) Operation {

	return OperationFunc(func(edge *graph.OpenEdge[gg.Value], scope Scope) (Value, error) {

		edgeVal, err := EvaluateEdge(edge, scope)

		if err != nil {
			return Value{}, err
		}

		return fn(edgeVal)
	})
}

// NOTE this is a giant hack to get around the fact that we're not yet
// using a genericized Graph implementation, so when we do AddValueIn
// on a graph.Graph[gg.Value] we can't use a Tuple value (because gg has no Tuple
// value), we have to use a Tuple vertex instead.
//
// This also doesn't yet support passing an operation as a value to another
// operation.
func preEvalEdgeOp(fn func(*graph.OpenEdge[gg.Value]) (Value, error)) Operation {

	return preEvalValOp(func(val Value) (Value, error) {

		var edge *graph.OpenEdge[gg.Value]

		if len(val.Tuple) > 0 {

			tupEdges := make([]*graph.OpenEdge[gg.Value], len(val.Tuple))

			for i := range val.Tuple {
				tupEdges[i] = graph.ValueOut[gg.Value](val.Tuple[i].Value, gg.ZeroValue)
			}

			edge = graph.TupleOut[gg.Value](tupEdges, gg.ZeroValue)

		} else {

			edge = graph.ValueOut[gg.Value](val.Value, gg.ZeroValue)

		}

		return fn(edge)
	})

}

type graphOp struct {
	*graph.Graph[gg.Value]
	scope Scope
}

// OperationFromGraph wraps the given Graph such that it can be used as an
// operation.
//
// When Perform is called the passed in OpenEdge is set to the "in" name value
// of the given Graph, then that resultant graph and the given parent Scope are
// used to construct a new Scope. The "out" name value is Evaluated on that
// Scope to obtain a resultant Value.
func OperationFromGraph(g *graph.Graph[gg.Value], scope Scope) Operation {
	return &graphOp{
		Graph: g,
		scope: scope,
	}
}

func (g *graphOp) Perform(edge *graph.OpenEdge[gg.Value], scope Scope) (Value, error) {

	return preEvalEdgeOp(func(edge *graph.OpenEdge[gg.Value]) (Value, error) {

		scope = ScopeFromGraph(
			g.Graph.AddValueIn(edge, inVal.Value),
			g.scope,
		)

		return scope.Evaluate(outVal)

	}).Perform(edge, scope)

}

// OperationFunc is a function which implements the Operation interface.
type OperationFunc func(*graph.OpenEdge[gg.Value], Scope) (Value, error)

// Perform calls the underlying OperationFunc directly.
func (f OperationFunc) Perform(edge *graph.OpenEdge[gg.Value], scope Scope) (Value, error) {
	return f(edge, scope)
}
