package vm

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/mediocregopher/ginger/lang"
	"llvm.org/llvm/bindings/go/llvm"
)

type op interface {
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

func (vt valType) eq(vt2 valType) bool {
	return lang.Equal(vt.term, vt2.term) && vt.llvm == vt2.llvm
}

// primitive valTypes
var (
	valTypeInt = valType{term: Int, llvm: llvm.Int64Type()}
)

////////////////////////////////////////////////////////////////////////////////

// most types don't have an input, so we use this as a shortcut
type voidIn struct{}

func (voidIn) inType() valType {
	return valType{
		term: lang.Tuple{},
		llvm: llvm.VoidType(),
	}
}

////////////////////////////////////////////////////////////////////////////////

type intOp struct {
	voidIn
	c lang.Const
}

func (io intOp) outType() valType {
	return valTypeInt
}

func (io intOp) build(mod *Module) (llvm.Value, error) {
	ci, err := strconv.ParseInt(string(io.c), 10, 64)
	if err != nil {
		return llvm.Value{}, err
	}
	return llvm.ConstInt(llvm.Int64Type(), uint64(ci), false), nil
}

////////////////////////////////////////////////////////////////////////////////

type tupOp struct {
	voidIn
	els []op
}

func (to tupOp) outType() valType {
	termTypes := make(lang.Tuple, len(to.els))
	llvmTypes := make([]llvm.Type, len(to.els))
	for i := range to.els {
		elValType := to.els[i].outType()
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

func (to tupOp) build(mod *Module) (llvm.Value, error) {
	str := llvm.Undef(to.outType().llvm)
	var val llvm.Value
	var err error
	for i := range to.els {
		if val, err = to.els[i].build(mod); err != nil {
			str = mod.b.CreateInsertValue(str, val, i, "")
		}
	}
	return str, err
}

////////////////////////////////////////////////////////////////////////////////

type addOp struct {
	voidIn
	a, b op
}

func (ao addOp) outType() valType {
	return ao.a.outType()
}

func (ao addOp) build(mod *Module) (llvm.Value, error) {
	av, err := ao.a.build(mod)
	if err != nil {
		return llvm.Value{}, err
	}
	bv, err := ao.b.build(mod)
	if err != nil {
		return llvm.Value{}, err
	}
	return mod.b.CreateAdd(av, bv, ""), nil
}

////////////////////////////////////////////////////////////////////////////////

func termToOp(t lang.Term) (op, error) {
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
	vAsTup := func(n int) ([]op, error) {
		vop, err := termToOp(v)
		if err != nil {
			return nil, err
		}
		vtup, ok := vop.(tupOp)
		if !ok || len(vtup.els) != n {
			return nil, fmt.Errorf("op %v expects a %d-tuple argument", k, n)
		}
		return vtup.els, nil
	}

	switch k {
	case Int:
		if !lang.Match(cPat(lang.AUnder), v) {
			return nil, errors.New("int requires constant arg")
		}
		return intOp{c: v.(lang.Const)}, nil
	case Tuple:
		if !lang.Match(lang.Tuple{Tuple, lang.AUnder}, v) {
			return nil, errors.New("tup requires tuple arg")
		}
		tup := v.(lang.Tuple)
		tc := tupOp{els: make([]op, len(tup))}
		var err error
		for i := range tup {
			if tc.els[i], err = termToOp(tup[i]); err != nil {
				return nil, err
			}
		}
		return tc, nil
	case Add:
		els, err := vAsTup(2)
		if err != nil {
			return nil, err
		} else if !els[0].outType().eq(valTypeInt) {
			return nil, errors.New("add args must be numbers of the same type")
		} else if !els[1].outType().eq(valTypeInt) {
			return nil, errors.New("add args must be numbers of the same type")
		}
		return addOp{a: els[0], b: els[1]}, nil
	default:
		return nil, fmt.Errorf("op %v unknown, or its args are malformed", t)
	}
}
