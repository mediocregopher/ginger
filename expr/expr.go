package expr

import (
	"fmt"

	"llvm.org/llvm/bindings/go/llvm"
)

// Expr represents the actual expression in question.
type Expr interface{}

// equaler is used to compare two expressions. The comparison should not take
// into account Token values, only the actual value being represented
type equaler interface {
	equal(equaler) bool
}

// will panic if either Expr doesn't implement equaler
func exprEqual(e1, e2 Expr) bool {
	eq1, ok1 := e1.(equaler)
	eq2, ok2 := e2.(equaler)
	if !ok1 || !ok2 {
		panic(fmt.Sprintf("can't compare %T and %T", e1, e2))
	}
	return eq1.equal(eq2)
}

////////////////////////////////////////////////////////////////////////////////

// an Expr which simply wraps an existing llvm.Value
type llvmVal llvm.Value

/*
func voidVal(lctx LLVMCtx) llvmVal {
	return llvmVal{lctx.B.CreateRetVoid()}
}
*/

////////////////////////////////////////////////////////////////////////////////

/*
// Void represents no data (size = 0)
type Void struct{}

func (v Void) equal(e equaler) bool {
	_, ok := e.(Void)
	return ok
}
*/

////////////////////////////////////////////////////////////////////////////////
/*
// Bool represents a true or false value
type Bool bool

func (b Bool) equal(e equaler) bool {
	bb, ok := e.(Bool)
	if !ok {
		return false
	}
	return bb == b
}
*/
////////////////////////////////////////////////////////////////////////////////

// Int represents an integer value
type Int int64

func (i Int) equal(e equaler) bool {
	ii, ok := e.(Int)
	return ok && ii == i
}

////////////////////////////////////////////////////////////////////////////////
/*
// String represents a string value
type String string

func (s String) equal(e equaler) bool {
	ss, ok := e.(String)
	if !ok {
		return false
	}
	return ss == s
}
*/
////////////////////////////////////////////////////////////////////////////////

// Identifier represents a binding to some other value which has been given a
// name
type Identifier string

func (id Identifier) equal(e equaler) bool {
	idid, ok := e.(Identifier)
	return ok && idid == id
}

func isIdentifier(e Expr) bool {
	_, ok := e.(Identifier)
	return ok
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

func (m Macro) equal(e equaler) bool {
	mm, ok := e.(Macro)
	return ok && m == mm
}

////////////////////////////////////////////////////////////////////////////////

// Tuple represents a fixed set of expressions which are interacted with as if
// they were a single value
type Tuple []Expr

// NewTuple returns a Tuple around the given list of Exprs
func NewTuple(ee ...Expr) Tuple {
	return Tuple(ee)
}

func (tup Tuple) String() string {
	return "(" + exprsJoin(tup) + ")"
}

func (tup Tuple) equal(e equaler) bool {
	tuptup, ok := e.(Tuple)
	return ok && exprsEqual(tup, tuptup)
}

////////////////////////////////////////////////////////////////////////////////

// List represents an ordered set of Exprs, all of the same type. A List's size
// does not affect its type signature, unlike a Tuple
type List []Expr

// NewList returns a List around the given list of Exprs
func NewList(ee ...Expr) List {
	return List(ee)
}

func (l List) String() string {
	return "[" + exprsJoin(l) + "]"
}

func (l List) equal(e equaler) bool {
	ll, ok := e.(List)
	return ok && exprsEqual(l, ll)
}

////////////////////////////////////////////////////////////////////////////////

// Statement represents an actual action which will be taken. The input value is
// used as the input to the pipe, and the output of the pipe is the output of
// the statement
type Statement struct {
	Op, Arg Expr
}

// NewStatement returns a Statement whose Op is the first Expr. If the given
// list is empty Arg will be nil, if its length is one Arg will be that single
// Expr, otherwise Arg will be a Tuple of the list
func NewStatement(e Expr, ee ...Expr) Statement {
	s := Statement{Op: e}
	if len(ee) > 1 {
		s.Arg = NewTuple(ee...)
	} else if len(ee) == 1 {
		s.Arg = ee[0]
	}
	return s
}

func (s Statement) String() string {
	return fmt.Sprintf("(%v %s)", s.Op, s.Arg)
}

func (s Statement) equal(e equaler) bool {
	ss, ok := e.(Statement)
	return ok && exprEqual(s.Op, ss.Op) && exprEqual(s.Arg, ss.Arg)
}

////////////////////////////////////////////////////////////////////////////////

// Block represents a set of statements which share a scope, i.e. If one
// statement binds a variable the rest of the statements in the block can use
// that variable
type Block struct {
	In    []Expr
	Stmts []Expr
	Out   []Expr
}

func (b Block) String() string {
	return fmt.Sprintf(
		"{[%s][%s][%s]}",
		exprsJoin(b.In),
		exprsJoin(b.Stmts),
		exprsJoin(b.Out),
	)
}

/*
func (b Block) LLVMVal(ctx *Ctx, lctx LLVMCtx) llvm.Value {
	name := randStr() // TODO make this based on token
	// TODO make these based on actual statements
	out := llvm.Int64Type()
	in := []llvm.Type{}
	fn := llvm.AddFunction(lctx.M, name, llvm.FunctionType(out, in, false))
	block := llvm.AddBasicBlock(fn, "entry")
	lctx.B.SetInsertPoint(block, block.FirstInstruction())

	var v llvm.Value
	for _, se := range b {
		v = se.Actual.LLVMVal(ctx, lctx)
	}
	// last v is used as return
	// TODO empty return
	lctx.B.CreateRet(v)
	return fn
}
*/

func (b Block) equal(e equaler) bool {
	bb, ok := e.(Block)
	return ok &&
		exprsEqual(b.In, bb.In) &&
		exprsEqual(b.Stmts, bb.Stmts) &&
		exprsEqual(b.Out, bb.Out)
}
