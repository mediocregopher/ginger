// Package vm implements the execution of gg.Graphs as programs.
package vm

import (
	"io"
	"fmt"
	"strings"

	"github.com/mediocregopher/ginger/gg"
	"github.com/mediocregopher/ginger/graph"
)

// ZeroValue is a Value with no fields set. It is equivalent to the 0-tuple.
var ZeroValue Value

// Value extends a gg.Value to include Operations and Tuples as a possible
// types.
type Value struct {
	gg.Value

	Operation
	Tuple []Value
}

// IsZero returns true if the Value is the zero value (aka the 0-tuple).
// LexerToken (within the gg.Value) is ignored for this check.
func (v Value) IsZero() bool {
	return v.Equal(ZeroValue)
}

// Equal returns true if the passed in Value is equivalent, ignoring the
// LexerToken on either Value.
//
// Will panic if the passed in v2 is not a Value from this package.
func (v Value) Equal(v2g graph.Value) bool {

	v2 := v2g.(Value)

	switch {

	case !v.Value.IsZero() || !v2.Value.IsZero():
		return v.Value.Equal(v2.Value)

	case v.Operation != nil || v2.Operation != nil:
		// for now we say that Operations can't be compared. This will probably
		// get revisted later.
		return false

	case len(v.Tuple) == len(v2.Tuple):

		for i := range v.Tuple {
			if !v.Tuple[i].Equal(v2.Tuple[i]) {
				return false
			}
		}

		return true

	default:

		// if both were the zero value then both tuples would have the same
		// length (0), which is covered by the previous check. So anything left
		// over must be tuples with differing lengths.
		return false
	}

}

func (v Value) String() string {

	switch {

	case v.Operation != nil:

		// We can try to get better strings for ops later
		return "<op>"

	case !v.Value.IsZero():
		return v.Value.String()

	default:

		// we consider zero value to be the 0-tuple

		strs := make([]string, len(v.Tuple))

		for i := range v.Tuple {
			strs[i] = v.Tuple[i].String()
		}

		return fmt.Sprintf("(%s)", strings.Join(strs, ", "))

	}

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

	return op.Perform(graph.ValueOut[gg.Value](input, gg.ZeroValue), scope)
}
