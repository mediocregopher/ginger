// This package describes ginger's base types and the interfaces covered by them
package types

import (
	"fmt"
)

// Elem is a generic type which can be used as a wrapper type for all ginger
// types, both base types and data structures
type Elem interface {
	
	// Returns whether one element is equal to another. Since all ginger values
	// are immutable, this must be a deep-equals check.
	Equal(Elem) bool
}

// Number can either be either an Int or a Float
type Number interface {
	Elem
}

// Wraps a go type like int, string, or []byte. GoType is a struct whose only
// field is an interface{}, so using a pointer to is not necessary. Just pass
// around the value type.
type GoType struct {
	V interface{}
}

func (g GoType) Equal(e Elem) bool {
	if g2, ok := e.(GoType); ok {
		return g.V == g2.V
	}
	return false
}

func (g GoType) String() string {
	return fmt.Sprintf("%v", g.V)
}
