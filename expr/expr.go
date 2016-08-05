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
// TODO LLVMVal -> LLVMBuild?

type LLVMCtx struct {
	B llvm.Builder
	M llvm.Module
}

// Actual represents the actual expression in question, and has certain
// properties. It is wrapped by Expr which also holds onto contextual
// information, like the token to which Actual was originally parsed from
type Actual interface {
	// Returns the llvm.Type which the expression accepts as an input, if any
	LLVMInType(ctx *Ctx) llvm.Type

	// Returns the llvm.Type which the expressions outputs
	LLVMOutType(ctx *Ctx) llvm.Type

	// Initializes an llvm.Value and returns it.
	LLVMVal(*Ctx, LLVMCtx) llvm.Value
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

// LLVMInType passes through to the method on the underlying Actual
func (e Expr) LLVMInType(ctx *Ctx) llvm.Type {
	return e.Actual.LLVMInType(ctx)
}

// LLVMOutType passes through to the method on the underlying Actual
func (e Expr) LLVMOutType(ctx *Ctx) llvm.Type {
	return e.Actual.LLVMOutType(ctx)
}

// LLVMVal passes its arguments to the underlying Actual instance. It caches the
// result, so if this is called multiple times the underlying one is only called
// the first time.
func (e Expr) LLVMVal(ctx *Ctx, lctx LLVMCtx) llvm.Value {
	if e.val == nil {
		v := e.Actual.LLVMVal(ctx, lctx)
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

// Void represents no data (size = 0)
type Void struct{}

// LLVMInType implements the Actual interface method
func (v Void) LLVMInType(ctx *Ctx) llvm.Type {
	panic("Void has no InType")
}

// LLVMOutType implements the Actual interface method
func (v Void) LLVMOutType(ctx *Ctx) llvm.Type {
	return llvm.VoidType()
}

// LLVMVal implements the Actual interface method
func (v Void) LLVMVal(ctx *Ctx, lctx LLVMCtx) llvm.Value {
	// Kind of weird that this only works for return type, but I guess makes
	// sense
	return lctx.B.CreateRetVoid()
}

////////////////////////////////////////////////////////////////////////////////

// Bool represents a true or false value
type Bool bool

// LLVMInType implements the Actual interface method
func (b Bool) LLVMInType(ctx *Ctx) llvm.Type {
	panic("Bool has no InType")
}

// LLVMOutType implements the Actual interface method
func (b Bool) LLVMOutType(ctx *Ctx) llvm.Type {
	return llvm.IntType(1)
}

// LLVMVal implements the Actual interface method
func (b Bool) LLVMVal(ctx *Ctx, lctx LLVMCtx) llvm.Value {
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

// LLVMInType implements the Actual interface method
func (i Int) LLVMInType(ctx *Ctx) llvm.Type {
	panic("Int has no InType")
}

// LLVMOutType implements the Actual interface method
func (i Int) LLVMOutType(ctx *Ctx) llvm.Type {
	return llvm.Int64Type()
}

// LLVMVal implements the Actual interface method
func (i Int) LLVMVal(ctx *Ctx, lctx LLVMCtx) llvm.Value {
	v := lctx.B.CreateAlloca(llvm.Int64Type(), "")
	lctx.B.CreateStore(llvm.ConstInt(llvm.Int64Type(), uint64(i), false), v)
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

// LLVMInType implements the Actual interface method
func (s String) LLVMInType(ctx *Ctx) llvm.Type {
	panic("String has no InType")
}

// LLVMOutType implements the Actual interface method
func (s String) LLVMOutType(ctx *Ctx) llvm.Type {
	panic("TODO")
}

// LLVMVal implements the Actual interface method
func (s String) LLVMVal(ctx *Ctx, lctx LLVMCtx) llvm.Value {
	panic("TODO")
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

// LLVMInType implements the Actual interface method
func (id Identifier) LLVMInType(ctx *Ctx) llvm.Type {
	panic("TODO")
}

// LLVMOutType implements the Actual interface method
func (id Identifier) LLVMOutType(ctx *Ctx) llvm.Type {
	panic("TODO")
}

// LLVMVal implements the Actual interface method
func (id Identifier) LLVMVal(ctx *Ctx, lctx LLVMCtx) llvm.Value {
	panic("TODO")
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

// LLVMInType implements the Actual interface method
func (m Macro) LLVMInType(ctx *Ctx) llvm.Type {
	panic("Macro has no InType")
}

// LLVMOutType implements the Actual interface method
func (m Macro) LLVMOutType(ctx *Ctx) llvm.Type {
	panic("Macro has no OutType")
}

// LLVMVal implements the Actual interface method
func (m Macro) LLVMVal(ctx *Ctx, lctx LLVMCtx) llvm.Value {
	panic("Macro has no Val")
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

// LLVMInType implements the Actual interface method
func (tup Tuple) LLVMInType(ctx *Ctx) llvm.Type {
	panic("TODO")
}

// LLVMOutType implements the Actual interface method
func (tup Tuple) LLVMOutType(ctx *Ctx) llvm.Type {
	panic("TODO")
}

// LLVMVal implements the Actual interface method
func (tup Tuple) LLVMVal(ctx *Ctx, lctx LLVMCtx) llvm.Value {
	panic("TODO")
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

func (s Statement) maybeMacro(ctx *Ctx) (Expr, bool) {
	m, ok := s.To.Actual.(Macro)
	if !ok {
		return Expr{}, false
	}

	fn := ctx.GetMacro(m)
	if fn == nil {
		return Expr{}, false
	}
	newe, err := fn(s.In)
	if err != nil {
		// TODO proper error
		panic(err)
	}
	return newe, true
}

// LLVMInType implements the Actual interface method
func (s Statement) LLVMInType(ctx *Ctx) llvm.Type {
	if newe, ok := s.maybeMacro(ctx); ok {
		return newe.LLVMInType(ctx)
	}
	// TODO futures
	panic("unknown Statement.To")
}

// LLVMOutType implements the Actual interface method
func (s Statement) LLVMOutType(ctx *Ctx) llvm.Type {
	if newe, ok := s.maybeMacro(ctx); ok {
		return newe.LLVMOutType(ctx)
	}
	panic("unknown Statement.To")
}

// LLVMVal implements the Actual interface method
func (s Statement) LLVMVal(ctx *Ctx, lctx LLVMCtx) llvm.Value {
	if newe, ok := s.maybeMacro(ctx); ok {
		return newe.LLVMVal(ctx, lctx)
	}
	panic("unknown Statement.To")
}

func (s Statement) equal(e equaler) bool {
	ss, ok := e.(Statement)
	return ok && s.In.equal(ss.In) && s.To.equal(ss.To)
}

////////////////////////////////////////////////////////////////////////////////

// Block represents a set of statements which share a scope, i.e. If one
// statement binds a variable the rest of the statements in the block can use
// that variable, including sub-blocks within this one.
type Block []Expr

func (b Block) String() string {
	strs := make([]string, len(b))
	for i := range b {
		strs[i] = b[i].Actual.(Statement).String()
	}
	return fmt.Sprintf("{ %s }", strings.Join(strs, " "))
}

// LLVMInType implements the Actual interface method
func (b Block) LLVMInType(ctx *Ctx) llvm.Type {
	panic("TODO")
}

// LLVMOutType implements the Actual interface method
func (b Block) LLVMOutType(ctx *Ctx) llvm.Type {
	return b[len(b)-1].LLVMOutType(ctx)
}

// LLVMVal implements the Actual interface method
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
