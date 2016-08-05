package expr

import (
	"fmt"

	"llvm.org/llvm/bindings/go/llvm"

	"github.com/mediocregopher/ginger/lexer"
)

// TODO empty blocks?
// TODO empty parenthesis
// TODO need to figure out how to test LLVMVal stuff
// TODO once we're a bit more confident, make ActualFunc

// Actual represents the actual expression in question. It is wrapped by Expr
// which also holds onto contextual information, like the token to which Actual
// was originally parsed from
type Actual interface {
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

// Void represents no data (size = 0)
type Void struct{}

func (v Void) equal(e equaler) bool {
	_, ok := e.(Void)
	return ok
}

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

func (i Int) build(lctx LLVMCtx) llvm.Value {
	v := lctx.B.CreateAlloca(llvm.Int64Type(), "")
	lctx.B.CreateStore(llvm.ConstInt(llvm.Int64Type(), uint64(i), false), v)
	return v
}

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

func (tup Tuple) String() string {
	return "(" + exprsJoin(tup) + ")"
}

func (tup Tuple) equal(e equaler) bool {
	tuptup, ok := e.(Tuple)
	return ok && exprsEqual(tup, tuptup)
}

////////////////////////////////////////////////////////////////////////////////

// Statement represents an actual action which will be taken. The input value is
// used as the input to the pipe, and the output of the pipe is the output of
// the statement
type Statement struct {
	Op, Arg Expr
}

func (s Statement) String() string {
	return fmt.Sprintf("(%v %s)", s.Op.Actual, s.Arg.Actual)
}

func (s Statement) equal(e equaler) bool {
	ss, ok := e.(Statement)
	return ok && s.Op.equal(ss.Op) && s.Arg.equal(ss.Arg)
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
