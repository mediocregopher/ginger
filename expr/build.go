package expr

import (
	"fmt"
	"log"

	"llvm.org/llvm/bindings/go/llvm"
)

func init() {
	log.Printf("initializing llvm")
	llvm.LinkInMCJIT()
	llvm.InitializeNativeTarget()
	llvm.InitializeNativeAsmPrinter()
}

type BuildCtx struct {
	B llvm.Builder
	M llvm.Module
}

func NewBuildCtx(moduleName string) BuildCtx {
	return BuildCtx{
		B: llvm.NewBuilder(),
		M: llvm.NewModule(moduleName),
	}
}

func (bctx BuildCtx) Build(ctx Ctx, stmts ...Statement) llvm.Value {
	var lastVal llvm.Value
	for _, stmt := range stmts {
		if e := bctx.BuildStmt(ctx, stmt); e != nil {
			if lv, ok := e.(llvmVal); ok {
				lastVal = llvm.Value(lv)
			} else {
				log.Printf("BuildStmt returned non llvmVal from %v: %v (%T)", stmt, e, e)
			}
		}
	}
	if (lastVal == llvm.Value{}) {
		lastVal = bctx.B.CreateRetVoid()
	}
	return lastVal
}

func (bctx BuildCtx) BuildStmt(ctx Ctx, s Statement) Expr {
	log.Printf("building: %v", s)
	switch o := s.Op.(type) {
	case Macro:
		return ctx.Macro(o)(bctx, ctx, s.Arg)
	case Identifier:
		s2 := s
		s2.Op = ctx.Identifier(o).(llvmVal)
		return bctx.BuildStmt(ctx, s2)
	case Statement:
		s2 := s
		s2.Op = bctx.BuildStmt(ctx, o)
		return bctx.BuildStmt(ctx, s2)
	case llvmVal:
		arg := bctx.buildExpr(ctx, s.Arg).(llvmVal)
		out := bctx.B.CreateCall(llvm.Value(o), []llvm.Value{llvm.Value(arg)}, "")
		return llvmVal(out)
	default:
		panic(fmt.Sprintf("non op type %v (%T)", s.Op, s.Op))
	}
}

// may return nil if e is a Statement which has no return
func (bctx BuildCtx) buildExpr(ctx Ctx, e Expr) Expr {
	return bctx.buildExprTill(ctx, e, func(Expr) bool { return false })
}

// like buildExpr, but will stop short and stop recursing when the function
// returns true
func (bctx BuildCtx) buildExprTill(ctx Ctx, e Expr, fn func(e Expr) bool) Expr {
	if fn(e) {
		return e
	}

	switch ea := e.(type) {
	case llvmVal:
		return e
	case Int:
		return llvmVal(llvm.ConstInt(llvm.Int64Type(), uint64(ea), false))
	case Identifier:
		return ctx.Identifier(ea)
	case Statement:
		return bctx.BuildStmt(ctx, ea)
	case Tuple:
		// if the tuple is empty then it is a void
		if len(ea) == 0 {
			return llvmVal(llvm.Undef(llvm.VoidType()))
		}

		ea2 := make(Tuple, len(ea))
		for i := range ea {
			ea2[i] = bctx.buildExprTill(ctx, ea[i], fn)
		}

		// if the fields of the tuple are all llvmVal then we can make a proper
		// struct
		vals := make([]llvm.Value, len(ea2))
		typs := make([]llvm.Type, len(ea2))
		for i := range ea2 {
			if v, ok := ea2[i].(llvmVal); ok {
				val := llvm.Value(v)
				vals[i] = val
				typs[i] = val.Type()
			} else {
				return ea2
			}
		}

		str := llvm.Undef(llvm.StructType(typs, false))
		for i := range vals {
			str = bctx.B.CreateInsertValue(str, vals[i], i, "")
		}
		return llvmVal(str)
	case List:
		ea2 := make(Tuple, len(ea))
		for i := range ea {
			ea2[i] = bctx.buildExprTill(ctx, ea[i], fn)
		}
		return ea2
	case Ctx:
		return ea
	default:
		panicf("%v (type %T) can't express a value", ea, ea)
	}
	panic("go is dumb")
}

func (bctx BuildCtx) buildVal(ctx Ctx, e Expr) llvm.Value {
	return llvm.Value(bctx.buildExpr(ctx, e).(llvmVal))
}

// globalCtx describes what's available to *all* contexts, and is what all
// contexts should have as the root parent in the tree.
//
// We define in this weird way cause NewCtx actually references globalCtx
var globalCtx *Ctx
var _ = func() bool {
	globalCtx = &Ctx{
		macros: map[Macro]MacroFn{
			"add": func(bctx BuildCtx, ctx Ctx, e Expr) Expr {
				tup := bctx.buildExpr(ctx, e).(llvmVal)
				a := bctx.B.CreateExtractValue(llvm.Value(tup), 0, "")
				b := bctx.B.CreateExtractValue(llvm.Value(tup), 1, "")
				return llvmVal(bctx.B.CreateAdd(a, b, ""))
			},

			// TODO this chould be a user macro!!!! WUT this language is baller
			"bind": func(bctx BuildCtx, ctx Ctx, e Expr) Expr {
				tup := bctx.buildExprTill(ctx, e, isIdentifier).(Tuple)
				id := bctx.buildExprTill(ctx, tup[0], isIdentifier).(Identifier)
				val := bctx.buildExpr(ctx, tup[1])
				ctx.Bind(id, val)
				return NewTuple()
			},

			"ctxnew": func(bctx BuildCtx, ctx Ctx, e Expr) Expr {
				return NewCtx()
			},

			"ctxthis": func(bctx BuildCtx, ctx Ctx, e Expr) Expr {
				return ctx
			},

			"ctxbind": func(bctx BuildCtx, ctx Ctx, e Expr) Expr {
				tup := bctx.buildExprTill(ctx, e, isIdentifier).(Tuple)
				thisCtx := bctx.buildExpr(ctx, tup[0]).(Ctx)
				id := bctx.buildExprTill(ctx, tup[1], isIdentifier).(Identifier)
				thisCtx.Bind(id, bctx.buildExpr(ctx, tup[2]))
				return NewTuple()
			},

			"ctxget": func(bctx BuildCtx, ctx Ctx, e Expr) Expr {
				tup := bctx.buildExprTill(ctx, e, isIdentifier).(Tuple)
				thisCtx := bctx.buildExpr(ctx, tup[0]).(Ctx)
				id := bctx.buildExprTill(ctx, tup[1], isIdentifier).(Identifier)
				return thisCtx.Identifier(id)
			},

			"do": func(bctx BuildCtx, ctx Ctx, e Expr) Expr {
				tup := bctx.buildExprTill(ctx, e, isStmt).(Tuple)
				thisCtx := tup[0].(Ctx)
				for _, stmtE := range tup[1].(List) {
					bctx.BuildStmt(thisCtx, stmtE.(Statement))
				}
				return NewTuple()
			},

			"op": func(bctx BuildCtx, ctx Ctx, e Expr) Expr {
				l := bctx.buildExprTill(ctx, e, isList).(List)
				stmts := make([]Statement, len(l))
				for i := range l {
					stmts[i] = l[i].(Statement)
				}

				// TODO obviously this needs to be fixed
				fn := llvm.AddFunction(bctx.M, "", llvm.FunctionType(llvm.Int64Type(), []llvm.Type{llvm.Int64Type()}, false))
				fnbl := llvm.AddBasicBlock(fn, "")

				prevbl := bctx.B.GetInsertBlock()
				bctx.B.SetInsertPoint(fnbl, fnbl.FirstInstruction())
				out := bctx.Build(NewCtx(), stmts...)
				bctx.B.CreateRet(out)
				bctx.B.SetInsertPointAtEnd(prevbl)
				return llvmVal(fn)
			},

			"in": func(bctx BuildCtx, ctx Ctx, e Expr) Expr {
				fn := bctx.B.GetInsertBlock().Parent()
				return llvmVal(fn.Param(0))
			},
		},
	}
	return false
}()
