package vm

import (
	"fmt"

	"github.com/mediocregopher/ginger/gg"
	"github.com/mediocregopher/ginger/graph"
)

// Operation is an entity which accepts an argument Value and performs some
// internal processing on that argument to return a resultant Value.
type Operation interface {
	Perform(Value) Value
}

// OperationFunc is a function which implements the Operation interface.
type OperationFunc func(Value) Value

// Perform calls the underlying OperationFunc directly.
func (f OperationFunc) Perform(arg Value) Value {
	return f(arg)
}

// Identity returns an Operation which always returns the given Value,
// regardless of the input argument.
//
// TODO this might not be the right name
func Identity(val Value) Operation {
	return OperationFunc(func(Value) Value {
		return val
	})
}

type graphOp struct {
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

// OperationFromGraph wraps the given Graph such that it can be used as an
// Operation. The given Scope determines what values outside of the Graph are
// available for use within the Operation.
func OperationFromGraph(g *gg.Graph, scope Scope) (Operation, error) {

	// edgeOp is distinct from a generic Operation in that the Value passed into
	// Perform will _always_ be the value of "in" for the overall Operation.
	//
	// edgeOps will wrap each other, passing "in" downwards to the leaf edgeOps.
	type edgeOp Operation

	var compileEdge func(*gg.OpenEdge) (edgeOp, error)

	// TODO memoize?
	valToEdgeOp := func(val Value) (edgeOp, error) {

		if val.Name == nil {
			return edgeOp(Identity(val)), nil
		}

		name := *val.Name

		if val.Equal(valNameIn) {
			return edgeOp(OperationFunc(func(inArg Value) Value {
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

			return edgeOp(Identity(val)), nil

		} else if l != 1 {
			return nil, fmt.Errorf("resolved name %q to %d input edges, rather than one", name, l)
		}

		edge := edgesIn[0]

		return compileEdge(edge)
	}

	// "out" resolves to more than a static value, treat the graph as a full
	// operation.

	// thisOp is used to support recur. It will get filled in with the Operation
	// which is returned by this function, once that Operation is created.
	thisOp := new(Operation)

	compileEdge = func(edge *gg.OpenEdge) (edgeOp, error) {

		return graph.MapReduce[gg.Value, gg.Value, edgeOp](
			edge,
			func(ggVal gg.Value) (edgeOp, error) {
				return valToEdgeOp(Value{Value: ggVal})
			},
			func(ggEdgeVal gg.Value, inEdgeOps []edgeOp) (edgeOp, error) {

				if ggEdgeVal.Equal(valNameIf.Value) {

					if len(inEdgeOps) != 3 {
						return nil, fmt.Errorf("'if' requires a 3-tuple argument")
					}

					return edgeOp(OperationFunc(func(inArg Value) Value {

						if pred := inEdgeOps[0].Perform(inArg); pred.Equal(valNumberZero) {
							return inEdgeOps[2].Perform(inArg)
						}

						return inEdgeOps[1].Perform(inArg)

					})), nil
				}

				// "if" statements (above) are the only case where we want the
				// input edges to remain separated, otherwise they should always
				// be combined into a single edge whose value is a tuple. Do
				// that here.

				inEdgeOp := inEdgeOps[0]

				if len(inEdgeOps) > 1 {
					inEdgeOp = edgeOp(OperationFunc(func(inArg Value) Value {
						tupVals := make([]Value, len(inEdgeOps))

						for i := range inEdgeOps {
							tupVals[i] = inEdgeOps[i].Perform(inArg)
						}

						return Tuple(tupVals...)
					}))
				}

				edgeVal := Value{Value: ggEdgeVal}

				if edgeVal.IsZero() {
					return inEdgeOp, nil
				}

				if edgeVal.Equal(valNameRecur) {
					return edgeOp(OperationFunc(func(inArg Value) Value {
						return (*thisOp).Perform(inEdgeOp.Perform(inArg))
					})), nil
				}

				if edgeVal.Graph != nil {

					opFromGraph, err := OperationFromGraph(
						edgeVal.Graph,
						scope.NewScope(),
					)

					if err != nil {
						return nil, fmt.Errorf("compiling graph to operation: %w", err)
					}

					edgeVal = Value{Operation: opFromGraph}
				}

				// The Operation is known at compile-time, so we can wrap it
				// directly into an edgeVal using the existing inEdgeOp as the
				// input.
				if edgeVal.Operation != nil {
					return edgeOp(OperationFunc(func(inArg Value) Value {
						return edgeVal.Operation.Perform(inEdgeOp.Perform(inArg))
					})), nil
				}

				// the edgeVal is not an Operation at compile time, and so
				// it must become one at runtime. We must resolve edgeVal to an
				// edgeOp as well (edgeEdgeOp), and then at runtime that is
				// given the inArg and (hopefully) the resultant Operation is
				// called.

				edgeEdgeOp, err := valToEdgeOp(edgeVal)

				if err != nil {
					return nil, err
				}

				return edgeOp(OperationFunc(func(inArg Value) Value {

					runtimeEdgeVal := edgeEdgeOp.Perform(inArg)

					if runtimeEdgeVal.Graph != nil {

						runtimeOp, err := OperationFromGraph(
							runtimeEdgeVal.Graph,
							scope.NewScope(),
						)

						if err != nil {
							panic(fmt.Sprintf("compiling graph to operation: %v", err))
						}

						runtimeEdgeVal = Value{Operation: runtimeOp}
					}

					if runtimeEdgeVal.Operation == nil {
						panic("edge value must be an operation")
					}

					return runtimeEdgeVal.Operation.Perform(inEdgeOp.Perform(inArg))

				})), nil
			},
		)
	}

	graphOp, err := valToEdgeOp(valNameOut)

	if err != nil {
		return nil, err
	}

	*thisOp = Operation(graphOp)

	return Operation(graphOp), nil
}
