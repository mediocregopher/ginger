package expr

import (
	"errors"

	"llvm.org/llvm/bindings/go/llvm"
)

type addActual []Expr

func (aa addActual) LLVMVal(builder llvm.Builder) llvm.Value {
	a := builder.CreateLoad(aa[0].LLVMVal(builder), "")
	for i := range aa[1:] {
		b := builder.CreateLoad(aa[i+1].LLVMVal(builder), "")
		a = builder.CreateAdd(a, b, "")
	}
	return a
}

var macros = map[Macro]func(Expr) (Expr, error){
	"add": func(e Expr) (Expr, error) {
		tup, ok := e.Actual.(Tuple)
		if !ok {
			// TODO proper error
			return Expr{}, errors.New("add only accepts a tuple")
		}
		return Expr{
			Actual: addActual(tup),
			Token:  e.Token,
		}, nil
	},
}
