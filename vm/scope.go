package vm

import (
	"fmt"
)

// Scope encapsulates a set of name->Value mappings.
type Scope interface {

	// Resolve accepts a name and returns an Value.
	Resolve(string) (Value, error)

	// NewScope returns a new Scope which sub-operations within this Scope
	// should use for themselves.
	NewScope() Scope
}

// ScopeMap implements the Scope interface.
type ScopeMap map[string]Value

var _ Scope = ScopeMap{}

// Resolve uses the given name as a key into the ScopeMap map, and
// returns the Operation held there for the key, if any.
func (m ScopeMap) Resolve(name string) (Value, error) {

	v, ok := m[name]

	if !ok {
		return Value{}, fmt.Errorf("%q not defined", name)
	}

	return v, nil
}

// NewScope returns the ScopeMap as-is.
func (m ScopeMap) NewScope() Scope {
	return m
}

type scopeWith struct {
	Scope // parent
	name  string
	val   Value
}

// ScopeWith returns a copy of the given Scope, except that evaluating the given
// name will always return the given Value.
func ScopeWith(scope Scope, name string, val Value) Scope {
	return &scopeWith{
		Scope: scope,
		name:  name,
		val:   val,
	}
}

func (s *scopeWith) Resolve(name string) (Value, error) {
	if name == s.name {
		return s.val, nil
	}
	return s.Scope.Resolve(name)
}
