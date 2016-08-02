package expr

import (
	"errors"

	"llvm.org/llvm/bindings/go/llvm"
)

type addActual []Expr

func (aa addActual) LLVMVal(ctx *Ctx, lctx LLVMCtx) llvm.Value {
	a := lctx.B.CreateLoad(aa[0].LLVMVal(ctx, lctx), "")
	for i := range aa[1:] {
		b := lctx.B.CreateLoad(aa[i+1].LLVMVal(ctx, lctx), "")
		a = lctx.B.CreateAdd(a, b, "")
	}
	return a
}

// RootCtx describes what's available to *all* contexts, and is what all
// contexts should have as the root parent in the tree
var RootCtx = &Ctx{
	Macros: map[Macro]func(Expr) (Expr, error){
		"add": func(e Expr) (Expr, error) {
			tup, ok := e.Actual.(Tuple)
			if !ok {
				// TODO proper error
				return Expr{}, errors.New("add only accepts a tuple")
			}
			// TODO check that it's a tuple of integers too
			return Expr{
				Actual: addActual(tup),
				Token:  e.Token,
			}, nil
		},
	},
}
