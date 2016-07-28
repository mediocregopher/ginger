package expr

import (
	"fmt"
	"strings"

	"llvm.org/llvm/bindings/go/llvm"

	"github.com/mediocregopher/ginger/lexer"
)

// TODO empty blocks?
// TODO empty parenthesis
// TODO need to figure out how to test LLVMVal stuff
// TODO once we're a bit more confident, make ActualFunc

// Actual represents the actual expression in question, and has certain
// properties. It is wrapped by Expr which also holds onto contextual
// information, like the token to which Actual was originally parsed from
type Actual interface {
	// Initializes an llvm.Value and returns it.
	LLVMVal(llvm.Builder) llvm.Value
}

// equaler is used to compare two expressions. The comparison should not take
// into account Token values, only the actual value being represented
type equaler interface {
	equal(equaler) bool
}

// Expr contains the actual expression as well as some contextual information
// wrapping it. Most interactions will be with this and not with the Actual
// directly.
type Expr struct {
	Actual Actual

	// Token is a nice-to-have, nothing will break if it's not there
	Token lexer.Token

	val *llvm.Value
}

// LLVMVal passes its arguments to the underlying Actual instance. It caches the
// result, so if this is called multiple times the underlying one is only called
// the first time.
func (e Expr) LLVMVal(builder llvm.Builder) llvm.Value {
	if e.val == nil {
		v := e.Actual.LLVMVal(builder)
		e.val = &v
	}
	return *e.val
}

// will panic if either Expr's Actual doesn't implement equaler
func (e Expr) equal(e2 Expr) bool {
	eq1, ok1 := e.Actual.(equaler)
	eq2, ok2 := e2.Actual.(equaler)
	if !ok1 || !ok2 {
		panic(fmt.Sprintf("can't compare %T and %T", e.Actual, e2.Actual))
	}
	return eq1.equal(eq2)
}

////////////////////////////////////////////////////////////////////////////////

// Bool represents a true or false value
type Bool bool

// LLVMVal implements the Actual interface method
func (b Bool) LLVMVal(builder llvm.Builder) llvm.Value {
	return llvm.Value{}
}

func (b Bool) equal(e equaler) bool {
	bb, ok := e.(Bool)
	if !ok {
		return false
	}
	return bb == b
}

////////////////////////////////////////////////////////////////////////////////

// Int represents an integer value
type Int int64

// LLVMVal implements the Actual interface method
func (i Int) LLVMVal(builder llvm.Builder) llvm.Value {
	v := builder.CreateAlloca(llvm.Int64Type(), "")
	builder.CreateStore(llvm.ConstInt(llvm.Int64Type(), uint64(i), false), v)
	return v
}

func (i Int) equal(e equaler) bool {
	ii, ok := e.(Int)
	if !ok {
		return false
	}
	return ii == i
}

////////////////////////////////////////////////////////////////////////////////

// String represents a string value
type String string

// LLVMVal implements the Actual interface method
func (s String) LLVMVal(builder llvm.Builder) llvm.Value {
	return llvm.Value{}
}

func (s String) equal(e equaler) bool {
	ss, ok := e.(String)
	if !ok {
		return false
	}
	return ss == s
}

////////////////////////////////////////////////////////////////////////////////

// Identifier represents a binding to some other value which has been given a
// name
type Identifier string

// LLVMVal implements the Actual interface method
func (id Identifier) LLVMVal(builder llvm.Builder) llvm.Value {
	return llvm.Value{}
}

func (id Identifier) equal(e equaler) bool {
	idid, ok := e.(Identifier)
	if !ok {
		return false
	}
	return idid == id
}

////////////////////////////////////////////////////////////////////////////////

// Macro is an identifier for a macro which can be used to transform
// expressions. The tokens for macros start with a '%', but the Macro identifier
// itself has that stripped off
type Macro string

// String returns the Macro with a '%' prepended to it
func (m Macro) String() string {
	return "%" + string(m)
}

// LLVMVal implements the Actual interface method
func (m Macro) LLVMVal(builder llvm.Builder) llvm.Value {
	panic("Macros have no inherent LLVMVal")
}

func (m Macro) equal(e equaler) bool {
	mm, ok := e.(Macro)
	if !ok {
		return false
	}
	return m == mm
}

////////////////////////////////////////////////////////////////////////////////

// Tuple represents a fixed set of expressions which are interacted with as if
// they were a single value
type Tuple []Expr

func (tup Tuple) String() string {
	strs := make([]string, len(tup))
	for i := range tup {
		strs[i] = fmt.Sprint(tup[i].Actual)
	}
	return "(" + strings.Join(strs, ", ") + ")"
}

// LLVMVal implements the Actual interface method
func (tup Tuple) LLVMVal(builder llvm.Builder) llvm.Value {
	return llvm.Value{}
}

func (tup Tuple) equal(e equaler) bool {
	tuptup, ok := e.(Tuple)
	if !ok || len(tuptup) != len(tup) {
		return false
	}
	for i := range tup {
		if !tup[i].equal(tuptup[i]) {
			return false
		}
	}
	return true
}

////////////////////////////////////////////////////////////////////////////////

// Statement represents an actual action which will be taken. The input value is
// used as the input to the pipe, and the output of the pipe is the output of
// the statement
type Statement struct {
	In Expr
	To Expr
}

func (s Statement) String() string {
	return fmt.Sprintf("(%v > %s)", s.In.Actual, s.To.Actual)
}

// LLVMVal implements the Actual interface method
func (s Statement) LLVMVal(builder llvm.Builder) llvm.Value {
	m, ok := s.To.Actual.(Macro)
	if !ok {
		// TODO proper error
		panic("statement To is not a macro")
	}
	fn, ok := macros[m]
	if !ok {
		// TODO proper error
		panic(fmt.Sprintf("unknown macro %q", m))
	}
	newe, err := fn(s.In)
	if err != nil {
		// TODO proper error
		panic(err)
	}
	return newe.LLVMVal(builder)
}

func (s Statement) equal(e equaler) bool {
	ss, ok := e.(Statement)
	return ok && s.In.equal(ss.In) && s.To.equal(ss.To)
}

////////////////////////////////////////////////////////////////////////////////

// Block represents a set of statements which share a scope, i.e. If one
// statement binds a variable the rest of the statements in the block can use
// that variable, including sub-blocks within this one.
type Block []Statement

func (b Block) String() string {
	strs := make([]string, len(b))
	for i := range b {
		strs[i] = b[i].String()
	}
	return fmt.Sprintf("{ %s }", strings.Join(strs, " "))
}

// LLVMVal implements the Actual interface method
func (b Block) LLVMVal(builder llvm.Builder) llvm.Value {
	return llvm.Value{}
}

func (b Block) equal(e equaler) bool {
	bb, ok := e.(Block)
	if !ok {
		return false
	}
	for i := range b {
		if !b[i].equal(bb[i]) {
			return false
		}
	}
	return true
}
