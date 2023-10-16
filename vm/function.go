package vm

import (
	"fmt"

	"github.com/mediocregopher/ginger/gg"
	"github.com/mediocregopher/ginger/graph"
)

// Function is an entity which accepts an argument Value and performs some
// internal processing on that argument to return a resultant Value.
type Function interface {
	Perform(Value) Value
}

// FunctionFunc is a function which implements the Function interface.
type FunctionFunc func(Value) Value

// Perform calls the underlying FunctionFunc directly.
func (f FunctionFunc) Perform(arg Value) Value {
	return f(arg)
}

// Identity returns an Function which always returns the given Value,
// regardless of the input argument.
//
// TODO this might not be the right name
func Identity(val Value) Function {
	return FunctionFunc(func(Value) Value {
		return val
	})
}

type graphFn struct {
	*gg.Graph
	scope Scope
}

var (
	valNameIn     = Value{Value: gg.Name("in")}
	valNameOut    = Value{Value: gg.Name("out")}
	valNameIf     = Value{Value: gg.Name("if")}
	valNameRecur  = Value{Value: gg.Name("recur")}
	valNumberZero = Value{Value: gg.Number(0)}
)

// FunctionFromGraph wraps the given Graph such that it can be used as an
// Function. The given Scope determines what values outside of the Graph are
// available for use within the Function.
func FunctionFromGraph(g *gg.Graph, scope Scope) (Function, error) {

	// edgeFn is distinct from a generic Function in that the Value passed into
	// Perform will _always_ be the value of "in" for the overall Function.
	//
	// edgeFns will wrap each other, passing "in" downwards to the leaf edgeFns.
	type edgeFn Function

	var compileEdge func(*gg.OpenEdge) (edgeFn, error)

	// TODO memoize?
	valToEdgeFn := func(val Value) (edgeFn, error) {

		if val.Name == nil {
			return edgeFn(Identity(val)), nil
		}

		name := *val.Name

		if val.Equal(valNameIn) {
			return edgeFn(FunctionFunc(func(inArg Value) Value {
				return inArg
			})), nil
		}

		// TODO intercept if and recur?

		edgesIn := g.ValueIns(val.Value)

		if l := len(edgesIn); l == 0 {

			val, err := scope.Resolve(name)

			if err != nil {
				return nil, fmt.Errorf("resolving name %q from the outer scope: %w", name, err)
			}

			return edgeFn(Identity(val)), nil

		} else if l != 1 {
			return nil, fmt.Errorf("resolved name %q to %d input edges, rather than one", name, l)
		}

		edge := edgesIn[0]

		return compileEdge(edge)
	}

	// "out" resolves to more than a static value, treat the graph as a full
	// operation.

	// thisFn is used to support recur. It will get filled in with the Function
	// which is returned by this function, once that Function is created.
	thisFn := new(Function)

	compileEdge = func(edge *gg.OpenEdge) (edgeFn, error) {

		return graph.MapReduce[gg.Value, gg.Value, edgeFn](
			edge,
			func(ggVal gg.Value) (edgeFn, error) {
				return valToEdgeFn(Value{Value: ggVal})
			},
			func(ggEdgeVal gg.Value, inEdgeFns []edgeFn) (edgeFn, error) {

				if ggEdgeVal.Equal(valNameIf.Value) {

					if len(inEdgeFns) != 3 {
						return nil, fmt.Errorf("'if' requires a 3-tuple argument")
					}

					return edgeFn(FunctionFunc(func(inArg Value) Value {

						if pred := inEdgeFns[0].Perform(inArg); pred.Equal(valNumberZero) {
							return inEdgeFns[2].Perform(inArg)
						}

						return inEdgeFns[1].Perform(inArg)

					})), nil
				}

				// "if" statements (above) are the only case where we want the
				// input edges to remain separated, otherwise they should always
				// be combined into a single edge whose value is a tuple. Do
				// that here.

				inEdgeFn := inEdgeFns[0]

				if len(inEdgeFns) > 1 {
					inEdgeFn = edgeFn(FunctionFunc(func(inArg Value) Value {
						tupVals := make([]Value, len(inEdgeFns))

						for i := range inEdgeFns {
							tupVals[i] = inEdgeFns[i].Perform(inArg)
						}

						return Tuple(tupVals...)
					}))
				}

				edgeVal := Value{Value: ggEdgeVal}

				if edgeVal.IsZero() {
					return inEdgeFn, nil
				}

				if edgeVal.Equal(valNameRecur) {
					return edgeFn(FunctionFunc(func(inArg Value) Value {
						return (*thisFn).Perform(inEdgeFn.Perform(inArg))
					})), nil
				}

				if edgeVal.Graph != nil {

					opFromGraph, err := FunctionFromGraph(
						edgeVal.Graph,
						scope.NewScope(),
					)

					if err != nil {
						return nil, fmt.Errorf("compiling graph to operation: %w", err)
					}

					edgeVal = Value{Function: opFromGraph}
				}

				// The Function is known at compile-time, so we can wrap it
				// directly into an edgeVal using the existing inEdgeFn as the
				// input.
				if edgeVal.Function != nil {
					return edgeFn(FunctionFunc(func(inArg Value) Value {
						return edgeVal.Function.Perform(inEdgeFn.Perform(inArg))
					})), nil
				}

				// the edgeVal is not an Function at compile time, and so
				// it must become one at runtime. We must resolve edgeVal to an
				// edgeFn as well (edgeEdgeFn), and then at runtime that is
				// given the inArg and (hopefully) the resultant Function is
				// called.

				edgeEdgeFn, err := valToEdgeFn(edgeVal)

				if err != nil {
					return nil, err
				}

				return edgeFn(FunctionFunc(func(inArg Value) Value {

					runtimeEdgeVal := edgeEdgeFn.Perform(inArg)

					if runtimeEdgeVal.Graph != nil {

						runtimeFn, err := FunctionFromGraph(
							runtimeEdgeVal.Graph,
							scope.NewScope(),
						)

						if err != nil {
							panic(fmt.Sprintf("compiling graph to operation: %v", err))
						}

						runtimeEdgeVal = Value{Function: runtimeFn}
					}

					if runtimeEdgeVal.Function == nil {
						panic("edge value must be an operation")
					}

					return runtimeEdgeVal.Function.Perform(inEdgeFn.Perform(inArg))

				})), nil
			},
		)
	}

	graphFn, err := valToEdgeFn(valNameOut)

	if err != nil {
		return nil, err
	}

	*thisFn = Function(graphFn)

	return Function(graphFn), nil
}
