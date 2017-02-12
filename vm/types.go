package vm

import (
	"fmt"

	"github.com/mediocregopher/ginger/lang"
	"llvm.org/llvm/bindings/go/llvm"
)

// Types supported by the vm in addition to those which are part of lang
var (
	Atom  = lang.AAtom
	Tuple = lang.ATuple
	Int   = lang.Atom("int")
)

var (
	tupPat = lang.Tuple{lang.ATuple, lang.Tuple{
		lang.Tuple{lang.AAtom, Tuple},
		lang.Tuple{lang.ATuple, lang.AUnder},
	}}
)

func termToType(t lang.Term) (llvm.Type, error) {
	switch {
	case lang.Equal(t, Int):
		return llvm.Int64Type(), nil
	case lang.Match(tupPat, t):
		tup := t.(lang.Tuple)[1].(lang.Tuple)
		if len(tup) == 0 {
			return llvm.VoidType(), nil
		}
		var err error
		typs := make([]llvm.Type, len(tup))
		for i := range tup {
			if typs[i], err = termToType(tup[i]); err != nil {
				return llvm.Type{}, err
			}
		}
		return llvm.StructType(typs, false), nil
	default:
		return llvm.Type{}, fmt.Errorf("type %v not supported", t)
	}
}
