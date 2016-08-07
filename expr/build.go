package expr

import "llvm.org/llvm/bindings/go/llvm"

type BuildCtx struct {
	C *Ctx
	B llvm.Builder
	M llvm.Module
}

func (bctx BuildCtx) Build(stmts ...Statement) llvm.Value {
	var lastVal llvm.Value
	for _, stmt := range stmts {
		if e := bctx.BuildStmt(stmt); e != nil {
			lastVal = bctx.buildVal(e)
		}
	}
	if (lastVal == llvm.Value{}) {
		lastVal = bctx.B.CreateRetVoid()
	}
	return lastVal
}

func (bctx BuildCtx) BuildStmt(s Statement) Expr {
	m := s.Op.(Macro)

	fn := bctx.C.GetMacro(m)
	if fn == nil {
		panicf("unknown macro: %q", m)
	}

	return fn(bctx, s.Arg)
}

// may return nil if e is a Statement which has no return
func (bctx BuildCtx) buildExpr(e Expr) Expr {
	return bctx.buildExprTill(e, func(Expr) bool { return false })
}

// like buildExpr, but will stop short and stop recursing when the function
// returns true
func (bctx BuildCtx) buildExprTill(e Expr, fn func(e Expr) bool) Expr {
	if fn(e) {
		return e
	}

	switch ea := e.(type) {
	case llvmVal:
		return e
	case Int:
		return llvmVal(llvm.ConstInt(llvm.Int64Type(), uint64(ea), false))
	case Identifier:
		if ev := bctx.C.GetIdentifier(ea); ev != nil {
			return ev
		}
		panicf("identifier %q not found", ea)
	case Statement:
		return bctx.BuildStmt(ea)
	case Tuple:
		for i := range ea {
			ea[i] = bctx.buildExprTill(ea[i], fn)
		}
		return ea
	default:
		panicf("type %T can't express a value", ea)
	}
	panic("go is dumb")
}

func (bctx BuildCtx) buildVal(e Expr) llvm.Value {
	return llvm.Value(bctx.buildExpr(e).(llvmVal))
}

// globalCtx describes what's available to *all* contexts, and is what all
// contexts should have as the root parent in the tree
var globalCtx = &Ctx{
	macros: map[Macro]MacroFn{
		"add": func(bctx BuildCtx, e Expr) Expr {
			tup := bctx.buildExpr(e).(Tuple)
			a := bctx.buildVal(tup[0])
			b := bctx.buildVal(tup[1])
			return llvmVal(bctx.B.CreateAdd(a, b, ""))
		},

		"bind": func(bctx BuildCtx, e Expr) Expr {
			tup := bctx.buildExprTill(e, isIdentifier).(Tuple)
			id := bctx.buildExprTill(tup[0], isIdentifier).(Identifier)
			if bctx.C.idents[id] != nil {
				panicf("identifier %q is already bound", id)
			}
			bctx.C.idents[id] = bctx.buildExpr(tup[1])
			return nil
		},
	},
}
