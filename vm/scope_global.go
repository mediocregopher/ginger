package vm

import (
	"fmt"

	"github.com/mediocregopher/ginger/gg"
)

// GlobalScope contains operations and values which are available from within
// any operation in a ginger program.
var GlobalScope = ScopeMap{

	"add": Value{Operation: preEvalValOp(func(val Value) (Value, error) {

		if len(val.Tuple) == 0 {
			return Value{}, fmt.Errorf("add requires a non-zero tuple of numbers as an argument")
		}

		var sum int64

		for _, tupVal := range val.Tuple {

			if tupVal.Number == nil {
				return Value{}, fmt.Errorf("add requires a non-zero tuple of numbers as an argument")
			}

			sum += *tupVal.Number
		}

		return Value{Value: gg.Value{Number: &sum}}, nil

	})},
}
