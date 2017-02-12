package vm

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/mediocregopher/ginger/lang"
	"llvm.org/llvm/bindings/go/llvm"
)

type cmd interface {
	inType() valType
	outType() valType
	build(*Module) (llvm.Value, error)
}

type valType struct {
	term lang.Term
	llvm llvm.Type
}

func (vt valType) isInt() bool {
	return lang.Equal(Int, vt.term)
}

// most types don't have an input, so we use this as a shortcut
type voidIn struct{}

func (voidIn) inType() valType {
	return valType{
		term: lang.Tuple{},
		llvm: llvm.VoidType(),
	}
}

////////////////////////////////////////////////////////////////////////////////

type intCmd struct {
	voidIn
	c lang.Const
}

func (ic intCmd) outType() valType {
	return valType{
		term: Int,
		llvm: llvm.Int64Type(),
	}
}

func (ic intCmd) build(mod *Module) (llvm.Value, error) {
	ci, err := strconv.ParseInt(string(ic.c), 10, 64)
	if err != nil {
		return llvm.Value{}, err
	}
	return llvm.ConstInt(llvm.Int64Type(), uint64(ci), false), nil
}

////////////////////////////////////////////////////////////////////////////////

type tupCmd struct {
	voidIn
	els []cmd
}

func (tc tupCmd) outType() valType {
	termTypes := make(lang.Tuple, len(tc.els))
	llvmTypes := make([]llvm.Type, len(tc.els))
	for i := range tc.els {
		elValType := tc.els[i].outType()
		termTypes[i] = elValType.term
		llvmTypes[i] = elValType.llvm
	}
	vt := valType{term: lang.Tuple{Tuple, termTypes}}
	if len(llvmTypes) == 0 {
		vt.llvm = llvm.VoidType()
	} else {
		vt.llvm = llvm.StructType(llvmTypes, false)
	}
	return vt
}

func (tc tupCmd) build(mod *Module) (llvm.Value, error) {
	str := llvm.Undef(tc.outType().llvm)
	var val llvm.Value
	var err error
	for i := range tc.els {
		if val, err = tc.els[i].build(mod); err != nil {
			str = mod.b.CreateInsertValue(str, val, i, "")
		}
	}
	return str, err
}

////////////////////////////////////////////////////////////////////////////////

type addCmd struct {
	voidIn
	a, b cmd
}

func (ac addCmd) outType() valType {
	return ac.a.outType()
}

func (ac addCmd) build(mod *Module) (llvm.Value, error) {
	av, err := ac.a.build(mod)
	if err != nil {
		return llvm.Value{}, err
	}
	bv, err := ac.b.build(mod)
	if err != nil {
		return llvm.Value{}, err
	}
	return mod.b.CreateAdd(av, bv, ""), nil
}

////////////////////////////////////////////////////////////////////////////////

func matchCmd(t lang.Term) (cmd, error) {
	aPat := func(a lang.Atom) lang.Tuple {
		return lang.Tuple{lang.AAtom, a}
	}
	cPat := func(t lang.Term) lang.Tuple {
		return lang.Tuple{lang.AConst, t}
	}
	tPat := func(el ...lang.Term) lang.Tuple {
		return lang.Tuple{Tuple, lang.Tuple(el)}
	}

	if !lang.Match(tPat(aPat(lang.AUnder), lang.TDblUnder), t) {
		return nil, fmt.Errorf("term %v does not look like a vm command", t)
	}
	k := t.(lang.Tuple)[0].(lang.Atom)
	v := t.(lang.Tuple)[1]

	// for when v is a Tuple argument, convenience function for casting
	vAsTup := func(n int) ([]cmd, error) {
		vcmd, err := matchCmd(v)
		if err != nil {
			return nil, err
		}
		vtup, ok := vcmd.(tupCmd)
		if !ok || len(vtup.els) != n {
			return nil, fmt.Errorf("cmd %v expects a %d-tuple argument", k, n)
		}
		return vtup.els, nil
	}

	switch k {
	case Int:
		if !lang.Match(cPat(lang.AUnder), v) {
			return nil, errors.New("int requires constant arg")
		}
		return intCmd{c: v.(lang.Const)}, nil
	case Tuple:
		if !lang.Match(lang.Tuple{Tuple, lang.AUnder}, v) {
			return nil, errors.New("tup requires tuple arg")
		}
		tup := v.(lang.Tuple)
		tc := tupCmd{els: make([]cmd, len(tup))}
		var err error
		for i := range tup {
			if tc.els[i], err = matchCmd(tup[i]); err != nil {
				return nil, err
			}
		}
		return tc, nil
	case Add:
		els, err := vAsTup(2)
		if err != nil {
			return nil, err
		} else if !els[0].outType().isInt() || !els[1].outType().isInt() {
			return nil, errors.New("add args must be numbers of the same type")
		}
		return addCmd{a: els[0], b: els[1]}, nil
	default:
		return nil, fmt.Errorf("cmd %v unknown, or its args are malformed", t)
	}
}
