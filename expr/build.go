package expr

import (
	"fmt"

	"llvm.org/llvm/bindings/go/llvm"
)

type BuildCtx struct {
	B llvm.Builder
	M llvm.Module
}

func (bctx BuildCtx) Build(ctx Ctx, stmts ...Statement) llvm.Value {
	var lastVal llvm.Value
	for _, stmt := range stmts {
		fmt.Println(stmt)
		if e := bctx.BuildStmt(ctx, stmt); e != nil {
			if lv, ok := e.(llvmVal); ok {
				lastVal = llvm.Value(lv)
			}
		}
	}
	if (lastVal == llvm.Value{}) {
		lastVal = bctx.B.CreateRetVoid()
	}
	return lastVal
}

func (bctx BuildCtx) BuildStmt(ctx Ctx, s Statement) Expr {
	m := s.Op.(Macro)
	return ctx.Macro(m)(bctx, ctx, s.Arg)
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
		for i := range ea {
			ea[i] = bctx.buildExprTill(ctx, ea[i], fn)
		}
		return ea
	case List:
		for i := range ea {
			ea[i] = bctx.buildExprTill(ctx, ea[i], fn)
		}
		return ea
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
				tup := bctx.buildExpr(ctx, e).(Tuple)
				a := bctx.buildVal(ctx, tup[0])
				b := bctx.buildVal(ctx, tup[1])
				return llvmVal(bctx.B.CreateAdd(a, b, ""))
			},

			// TODO this chould be a user macro!!!! WUT this language is baller
			"bind": func(bctx BuildCtx, ctx Ctx, e Expr) Expr {
				tup := bctx.buildExprTill(ctx, e, isIdentifier).(Tuple)
				id := bctx.buildExprTill(ctx, tup[0], isIdentifier).(Identifier)
				ctx.Bind(id, bctx.buildExpr(ctx, tup[1]))
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
		},
	}
	return false
}()
