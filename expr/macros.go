package expr

import "llvm.org/llvm/bindings/go/llvm"

// RootCtx describes what's available to *all* contexts, and is what all
// contexts should have as the root parent in the tree
var RootCtx = &Ctx{
	Macros: map[Macro]MacroFn{
		"add": func(ctx *Ctx, lctx LLVMCtx, e Expr) (llvm.Value, bool) {
			tup := e.Actual.(Tuple)
			buildInt := func(e Expr) llvm.Value {
				return lctx.B.CreateLoad(e.Actual.(Int).build(lctx), "")
			}

			a := buildInt(tup[0])
			for i := range tup[1:] {
				b := buildInt(tup[i+1])
				a = lctx.B.CreateAdd(a, b, "")
			}
			return a, true
		},
	},
}
