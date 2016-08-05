package expr

import "llvm.org/llvm/bindings/go/llvm"

type LLVMCtx struct {
	B llvm.Builder
	M llvm.Module
}

func Build(ctx *Ctx, lctx LLVMCtx, stmts []Expr) {
	for _, stmt := range stmts {
		BuildStmt(ctx, lctx, stmt)
	}
}

func BuildStmt(ctx *Ctx, lctx LLVMCtx, stmtE Expr) {
	s := stmtE.Actual.(Statement)
	m := s.Op.Actual.(Macro)

	fn := ctx.GetMacro(m)
	if fn == nil {
		panicf("unknown macro: %q", m)
	}
	lv, ok := fn(ctx, lctx, s.Arg)
	if ok {
		ctx.LastVal = lv
	}
}
