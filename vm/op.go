package vm

import (
	"github.com/mediocregopher/ginger/gg"
)

var (
	outVal = nameVal("out")
)

// Thunk is returned from the performance of an Operation. When called it will
// return the result of that Operation having been called with the particular
// arguments which were passed in.
type Thunk func() (Value, error)

func valThunk(val Value) Thunk {
	return func() (Value, error) { return val, nil }
}

// evalThunks is used to coalesce the results of multiple Thunks into a single
// Thunk which will return a tuple Value. As a special case, if only one Thunk
// is given then it is returned directly (1-tuple is equivalent to its only
// element).
func evalThunks(args []Thunk) Thunk {

	if len(args) == 1 {
		return args[0]
	}

	return func() (Value, error) {

		var (
			err     error
			tupVals = make([]Value, len(args))
		)

		for i := range args {
			if tupVals[i], err = args[i](); err != nil {
				return ZeroValue, err
			}
		}

		return Value{Tuple: tupVals}, nil
	}
}

// Operation is an entity which can accept one or more arguments (each not
// having been evaluated yet) and return a Thunk which will perform some
// internal processing on those arguments and return a resultant Value.
//
// The Operation passed into Perform is the Operation which is calling the
// Perform. It may be nil.
type Operation interface {
	Perform([]Thunk, Operation) (Thunk, error)
}

func preEvalValOp(fn func(Value) (Value, error)) Operation {

	return OperationFunc(func(args []Thunk, _ Operation) (Thunk, error) {

		return func() (Value, error) {

			val, err := evalThunks(args)()

			if err != nil {
				return ZeroValue, err
			}

			return fn(val)

		}, nil

	})
}

type graphOp struct {
	*gg.Graph
	scope Scope
}

// OperationFromGraph wraps the given Graph such that it can be used as an
// operation.
//
// The Thunk returned by Perform will evaluate the passed in Thunks, and set
// them to the "in" name value of the given Graph. The "out" name value is
// Evaluated using the given Scope to obtain a resultant Value.
func OperationFromGraph(g *gg.Graph, scope Scope) Operation {
	return &graphOp{
		Graph: g,
		scope: scope,
	}
}

func (g *graphOp) Perform(args []Thunk, _ Operation) (Thunk, error) {
	return ScopeFromGraph(
		g.Graph,
		evalThunks(args),
		g.scope,
		g,
	).Evaluate(outVal)
}

// OperationFunc is a function which implements the Operation interface.
type OperationFunc func([]Thunk, Operation) (Thunk, error)

// Perform calls the underlying OperationFunc directly.
func (f OperationFunc) Perform(args []Thunk, op Operation) (Thunk, error) {
	return f(args, op)
}
