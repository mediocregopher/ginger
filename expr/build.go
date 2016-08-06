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

	return fn(bctx, bctx.buildExpr(s.Arg))
}

// may return nil if e is a Statement which has no return
func (bctx BuildCtx) buildExpr(e Expr) Expr {
	switch ea := e.(type) {
	case llvmVal:
		return e
	case Int:
		return llvmVal(llvm.ConstInt(llvm.Int64Type(), uint64(ea), false))
	case Identifier:
		return bctx.buildExpr(bctx.C.GetIdentifier(ea))
	case Statement:
		return bctx.BuildStmt(ea)
	case Tuple:
		for i := range ea {
			ea[i] = bctx.buildExpr(ea[i])
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
			tup := e.(Tuple)
			a := bctx.buildVal(tup[0])
			b := bctx.buildVal(tup[1])
			return llvmVal(bctx.B.CreateAdd(a, b, ""))
		},

		"bind": func(bctx BuildCtx, e Expr) Expr {
			tup := e.(Tuple)
			id := bctx.buildExpr(tup[0]).(Identifier)
			bctx.C.idents[id] = bctx.buildVal(tup[1])
			return nil
		},
	},
}
