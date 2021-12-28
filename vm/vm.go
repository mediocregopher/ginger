// Package vm implements the execution of gg.Graphs as programs.
package vm

import (
	"io"

	"github.com/mediocregopher/ginger/gg"
)

// Value extends a gg.Value to include Operations and Tuples as a possible
// types.
type Value struct {
	gg.Value

	Operation
	Tuple []Value
}

func nameVal(n string) Value {
	var val Value
	val.Name = &n
	return val
}

// EvaluateSource reads and parses the io.Reader as an operation, input is used
// as the argument to the operation, and the resultant value is returned.
//
// scope contains pre-defined operations and values which are available during
// the evaluation.
func EvaluateSource(opSrc io.Reader, input gg.Value, scope Scope) (Value, error) {
	lexer := gg.NewLexer(opSrc)

	g, err := gg.DecodeLexer(lexer)
	if err != nil {
		return Value{}, err
	}

	op := OperationFromGraph(g, scope.NewScope())

	return op.Perform(gg.ValueOut(input, gg.ZeroValue), scope)
}
