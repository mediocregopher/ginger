// Package gg implements graph serialization to/from the gg text format.
package gg

import (
	"fmt"

	"github.com/mediocregopher/ginger/graph"
)

// ZeroValue is a Value with no fields set.
var ZeroValue Value

// Value represents a value which can be serialized by the gg text format.
type Value struct {

	// Only one of these fields may be set
	Name   *string
	Number *int64
	Graph  *Graph

	// TODO coming soon!
	// String *string

	// Optional fields indicating the token which was used to construct this
	// Value, if any.
	LexerToken *LexerToken
}

// Name returns a name Value.
func Name(name string) Value {
	return Value{Name: &name}
}

// Number returns a number Value.
func Number(n int64) Value {
	return Value{Number: &n}
}

// IsZero returns true if the Value is the zero value (none of the sub-value
// fields are set). LexerToken is ignored for this check.
func (v Value) IsZero() bool {
	return v.Equal(ZeroValue)
}

// Equal returns true if the passed in Value is equivalent, ignoring the
// LexerToken on either Value.
//
// Will panic if the passed in v2 is not a Value from this package.
func (v Value) Equal(v2g graph.Value) bool {

	v2 := v2g.(Value)

	v.LexerToken, v2.LexerToken = nil, nil

	switch {

	case v == ZeroValue && v2 == ZeroValue:
		return true

	case v.Name != nil && v2.Name != nil && *v.Name == *v2.Name:
		return true

	case v.Number != nil && v2.Number != nil && *v.Number == *v2.Number:
		return true

	case v.Graph != nil && v2.Graph != nil && v.Graph.Equal(v2.Graph):
		return true

	default:
		return false
	}
}

func (v Value) String() string {

	switch {

	case v.Name != nil:
		return *v.Name

	case v.Number != nil:
		return fmt.Sprint(*v.Number)

	case v.Graph != nil:
		return v.Graph.String()

	default:
		return "<zero>"
	}
}
