package vm

import (
	"strconv"

	"github.com/mediocregopher/ginger/lang"
	"llvm.org/llvm/bindings/go/llvm"
)

type buildCmd struct {
	pattern   lang.Tuple
	inTypeFn  func(lang.Term) (llvm.Type, error)
	outTypeFn func(lang.Term) (llvm.Type, error)
	buildFn   func(lang.Term) (llvm.Value, error)
}

func (cmd buildCmd) matches(t lang.Term) bool {
	return lang.Match(cmd.pattern, t)
}

func (cmd buildCmd) inType(t lang.Term) (llvm.Type, error) {
	if cmd.inTypeFn == nil {
		return llvm.VoidType(), nil
	}
	return cmd.inTypeFn(t)
}

func (cmd buildCmd) outType(t lang.Term) (llvm.Type, error) {
	if cmd.outTypeFn == nil {
		return llvm.VoidType(), nil
	}
	return cmd.outTypeFn(t)
}

func (cmd buildCmd) build(t lang.Term) (llvm.Value, error) {
	return cmd.buildFn(t)
}

func buildCmds(mod *Module) []buildCmd {
	aPat := func(a lang.Atom) lang.Tuple {
		return lang.Tuple{lang.AAtom, a}
	}
	cPat := func(t lang.Term) lang.Tuple {
		return lang.Tuple{lang.AConst, t}
	}
	tPat := func(el ...lang.Term) lang.Tuple {
		return lang.Tuple{lang.ATuple, lang.Tuple(el)}
	}
	buildPat := func(a lang.Atom, b lang.Tuple) lang.Tuple {
		return tPat(aPat(a), b)
	}
	return []buildCmd{
		{ // (int 42)
			pattern: buildPat(lang.AInt, cPat(lang.AUnder)),
			outTypeFn: func(t lang.Term) (llvm.Type, error) {
				return llvm.Int64Type(), nil
			},
			buildFn: func(t lang.Term) (llvm.Value, error) {
				con := t.(lang.Const)
				coni, err := strconv.ParseInt(string(con), 10, 64)
				if err != nil {
					return llvm.Value{}, err
				}
				// TODO why does this have to be cast?
				return llvm.ConstInt(llvm.Int64Type(), uint64(coni), false), nil
			},
		},

		{ // (tup ((atom foo) (const 10)))
			pattern: buildPat(lang.ATuple, lang.Tuple{lang.ATuple, lang.AUnder}),
			outTypeFn: func(t lang.Term) (llvm.Type, error) {
				tup := t.(lang.Tuple)
				if len(tup) == 0 {
					return llvm.VoidType(), nil
				}
				var err error
				typs := make([]llvm.Type, len(tup))
				for i := range tup {
					if typs[i], err = mod.outType(tup[i]); err != nil {
						return llvm.Type{}, err
					}
				}
				return llvm.StructType(typs, false), nil
			},
			buildFn: func(t lang.Term) (llvm.Value, error) {
				tup := t.(lang.Tuple)
				// if the tuple is empty then it is a void
				if len(tup) == 0 {
					return llvm.Undef(llvm.VoidType()), nil
				}

				var err error
				vals := make([]llvm.Value, len(tup))
				typs := make([]llvm.Type, len(tup))
				for i := range tup {
					if vals[i], err = mod.build(tup[i]); err != nil {
						return llvm.Value{}, err
					}
					typs[i] = vals[i].Type()
				}

				str := llvm.Undef(llvm.StructType(typs, false))
				for i := range vals {
					str = mod.b.CreateInsertValue(str, vals[i], i, "")
				}
				return str, nil
			},
		},

		{ // (add ((const 5) (var foo)))
			pattern: buildPat(lang.AAdd, tPat(lang.TDblUnder, lang.TDblUnder)),
			outTypeFn: func(t lang.Term) (llvm.Type, error) {
				return llvm.Int64Type(), nil
			},
			buildFn: func(t lang.Term) (llvm.Value, error) {
				tup := t.(lang.Tuple)
				v1, err := mod.build(tup[0])
				if err != nil {
					return llvm.Value{}, err
				}
				v2, err := mod.build(tup[1])
				if err != nil {
					return llvm.Value{}, err
				}
				return mod.b.CreateAdd(v1, v2, ""), nil
			},
		},
	}
}
