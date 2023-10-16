package vm

import (
	"fmt"

	"github.com/mediocregopher/ginger/gg"
)

func globalFn(fn func(Value) (Value, error)) Value {
	return Value{
		Function: FunctionFunc(func(in Value) Value {
			res, err := fn(in)
			if err != nil {
				panic(err)
			}
			return res
		}),
	}
}

// GlobalScope contains operations and values which are available from within
// any operation in a ginger program.
var GlobalScope = ScopeMap{

	"add": globalFn(func(val Value) (Value, error) {

		var sum int64

		for _, tupVal := range val.Tuple {

			if tupVal.Number == nil {
				return Value{}, fmt.Errorf("add requires a non-zero tuple of numbers as an argument")
			}

			sum += *tupVal.Number
		}

		return Value{Value: gg.Value{Number: &sum}}, nil

	}),

	"tupEl": globalFn(func(val Value) (Value, error) {

		tup, i := val.Tuple[0], val.Tuple[1]

		return tup.Tuple[int(*i.Number)], nil

	}),

	"isZero": globalFn(func(val Value) (Value, error) {

		if *val.Number == 0 {
			one := int64(1)
			return Value{Value: gg.Value{Number: &one}}, nil
		}

		zero := int64(0)
		return Value{Value: gg.Value{Number: &zero}}, nil

	}),
}
