package lang

import (
	"fmt"
	"reflect"
	"strings"
)

// Term is a unit of language which carries some meaning. Some Terms are
// actually comprised of multiple sub-Terms.
type Term interface {
	fmt.Stringer // for debugging

	// Type returns a Term which describes the type of this Term, i.e.  the
	// components this Term is comprised of.
	Type() Term
}

// Equal returns whether or not two Terms are of equal value
func Equal(t1, t2 Term) bool {
	return reflect.DeepEqual(t1, t2)
}

////////////////////////////////////////////////////////////////////////////////

// Atom is a constant with no other meaning than that it can be equal or not
// equal to another Atom.
type Atom string

func (a Atom) String() string {
	return string(a)
}

// Type implements the method for Term
func (a Atom) Type() Term {
	return Atom("atom")
}

////////////////////////////////////////////////////////////////////////////////

// Const is a constant whose meaning depends on the context in which it is used
type Const string

func (a Const) String() string {
	return string(a)
}

// Type implements the method for Term
func (a Const) Type() Term {
	return Const("const")
}

////////////////////////////////////////////////////////////////////////////////

// Tuple is a compound Term of zero or more sub-Terms, each of which may have a
// different Type. Both the length of the Tuple and the Type of each of it's
// sub-Terms are components in the Tuple's Type.
type Tuple []Term

func (t Tuple) String() string {
	ss := make([]string, len(t))
	for i := range t {
		ss[i] = t[i].String()
	}
	return "(" + strings.Join(ss, " ") + ")"
}

// Type implements the method for Term
func (t Tuple) Type() Term {
	tt := make(Tuple, len(t))
	for i := range t {
		tt[i] = t[i].Type()
	}
	return Tuple{Atom("tup"), tt}
}

////////////////////////////////////////////////////////////////////////////////

type list struct {
	typ Term
	ll  []Term
}

// List is a compound Term of zero or more sub-Terms, each of which must have
// the same Type (the one given as the first argument to this function). Only
// the Type of the sub-Terms is a component in the List's Type.
func List(typ Term, elems ...Term) Term {
	return list{
		typ: typ,
		ll:  elems,
	}
}

func (l list) String() string {
	ss := make([]string, len(l.ll))
	for i := range l.ll {
		ss[i] = l.ll[i].String()
	}
	return "[" + strings.Join(ss, " ") + "]"
}

// Type implements the method for Term
func (l list) Type() Term {
	return Tuple{Atom("list"), l.typ}
}
