package vm

import (
	"fmt"

	"github.com/mediocregopher/ginger/gg"
)

// GlobalScope contains operations and values which are available from within
// any operation in a ginger program.
var GlobalScope = ScopeMap{

	"add": Value{Operation: preEvalValOp(func(val Value) (Value, error) {

		var sum int64

		for _, tupVal := range val.Tuple {

			if tupVal.Number == nil {
				return Value{}, fmt.Errorf("add requires a non-zero tuple of numbers as an argument")
			}

			sum += *tupVal.Number
		}

		return Value{Value: gg.Value{Number: &sum}}, nil

	})},

	"tupEl": Value{Operation: preEvalValOp(func(val Value) (Value, error) {

		tup, i := val.Tuple[0], val.Tuple[1]

		return tup.Tuple[int(*i.Number)], nil

	})},

	"isZero": Value{Operation: preEvalValOp(func(val Value) (Value, error) {

		if *val.Number == 0 {
			one := int64(1)
			return Value{Value: gg.Value{Number: &one}}, nil
		}

		zero := int64(0)
		return Value{Value: gg.Value{Number: &zero}}, nil

	})},

	"if": Value{Operation: OperationFunc(func(args []Thunk, _ Operation) (Thunk, error) {

		b := args[0]
		onTrue := args[1]
		onFalse := args[2]

		return func() (Value, error) {

			bVal, err := b()

			if err != nil {
				return ZeroValue, err
			}

			if *bVal.Number == 0 {
				return onFalse()
			}

			return onTrue()

		}, nil

	})},

	"recur": Value{Operation: OperationFunc(func(args []Thunk, op Operation) (Thunk, error) {
		return op.Perform(args, op)
	})},
}
