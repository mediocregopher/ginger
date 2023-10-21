// Package gg implements graph serialization to/from the gg text format.
package gg

import (
	"fmt"

	"github.com/mediocregopher/ginger/graph"
)

// Type aliases for convenience
type (
	Graph    = graph.Graph[OptionalValue, Value]
	OpenEdge = graph.OpenEdge[OptionalValue, Value]
)

// Value represents a value which can be serialized by the gg text format.
type Value struct {
	Location

	// Only one of these fields may be set
	Name   *string
	Number *int64
	Graph  *Graph
}

// Name returns a name Value.
func Name(name string) Value {
	return Value{Name: &name}
}

// Number returns a number Value.
func Number(n int64) Value {
	return Value{Number: &n}
}

// Equal returns true if the passed in Value is equivalent, ignoring the
// LexerToken on either Value.
//
// Will panic if the passed in v2 is not a Value from this package.
func (v Value) Equal(v2g graph.Value) bool {

	v2 := v2g.(Value)

	switch {

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
		panic("no fields set on Value")
	}
}

// OptionalValue is a Value which may be unset. This is used for edge values,
// since edges might not have a value.
type OptionalValue struct {
	Value
	Valid bool
}

// None is the zero OptionalValue (hello rustaceans).
var None OptionalValue

// Some wraps a Value to be an OptionalValue.
func Some(v Value) OptionalValue {
	return OptionalValue{Valid: true, Value: v}
}

func (v OptionalValue) String() string {
	if !v.Valid {
		return "<none>"
	}
	return v.Value.String()
}

func (v OptionalValue) Equal(v2g graph.Value) bool {
	var v2 OptionalValue

	if v2Val, ok := v2g.(Value); ok {
		v2 = Some(v2Val)
	} else {
		v2 = v2g.(OptionalValue)
	}

	if v.Valid != v2.Valid {
		return false
	} else if !v.Valid {
		return true
	}

	return v.Value.Equal(v2.Value)
}
