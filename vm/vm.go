package vm

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/mediocregopher/ginger/lang"

	"llvm.org/llvm/bindings/go/llvm"
)

// Val holds onto a value which has been created within the VM
type Val struct {
	v llvm.Value
}

// Module contains a compiled set of code which can be run, dumped in IR form,
// or compiled. A Module should be Dispose()'d of once it's no longer being
// used.
type Module struct {
	b      llvm.Builder
	m      llvm.Module
	mainFn llvm.Value
}

var initOnce sync.Once

// Build creates a new Module by compiling the given Term as code
func Build(t lang.Term) (*Module, error) {
	initOnce.Do(func() {
		llvm.LinkInMCJIT()
		llvm.InitializeNativeTarget()
		llvm.InitializeNativeAsmPrinter()
	})
	mod := &Module{
		b: llvm.NewBuilder(),
		m: llvm.NewModule(""),
	}

	// TODO figure out types
	mod.mainFn = llvm.AddFunction(mod.m, "", llvm.FunctionType(llvm.Int64Type(), []llvm.Type{}, false))
	block := llvm.AddBasicBlock(mod.mainFn, "")
	mod.b.SetInsertPoint(block, block.FirstInstruction())

	out, err := mod.build(t)
	if err != nil {
		mod.Dispose()
		return nil, err
	}
	mod.b.CreateRet(out.v)

	if err := llvm.VerifyModule(mod.m, llvm.ReturnStatusAction); err != nil {
		mod.Dispose()
		return nil, fmt.Errorf("could not verify module: %s", err)
	}

	return mod, nil
}

// Dispose cleans up all resources held by the Module
func (mod *Module) Dispose() {
	// TODO this panics for some reason...
	//mod.m.Dispose()
	//mod.b.Dispose()
}

func (mod *Module) build(t lang.Term) (Val, error) {
	aPat := func(a lang.Atom) lang.Tuple {
		return lang.Tuple{lang.AAtom, a}
	}
	cPat := func(t lang.Term) lang.Tuple {
		return lang.Tuple{lang.AConst, t}
	}
	tPat := func(el ...lang.Term) lang.Tuple {
		return lang.Tuple{lang.ATuple, lang.Tuple(el)}
	}
	match := func(a lang.Atom, b lang.Tuple) bool {
		return lang.Match(tPat(aPat(a), b), t)
	}

	switch {
	// (int 42)
	case match(lang.AInt, cPat(lang.AUnder)):
		con := t.(lang.Tuple)[1].(lang.Const)
		coni, err := strconv.ParseInt(string(con), 10, 64)
		if err != nil {
			return Val{}, err
		}
		return Val{
			// TODO why does this have to be cast?
			v: llvm.ConstInt(llvm.Int64Type(), uint64(coni), false),
		}, nil

	// (tup ((atom foo) (const 10)))
	case match(lang.ATuple, lang.Tuple{lang.ATuple, lang.AUnder}):
		tup := t.(lang.Tuple)[1].(lang.Tuple)
		// if the tuple is empty then it is a void
		if len(tup) == 0 {
			return Val{v: llvm.Undef(llvm.VoidType())}, nil
		}

		var err error
		vals := make([]Val, len(tup))
		typs := make([]llvm.Type, len(tup))
		for i := range tup {
			if vals[i], err = mod.build(tup[i]); err != nil {
				return Val{}, err
			}
			typs[i] = vals[i].v.Type()
		}

		str := llvm.Undef(llvm.StructType(typs, false))
		for i := range vals {
			str = mod.b.CreateInsertValue(str, vals[i].v, i, "")
		}
		return Val{v: str}, nil

	// (add ((const 5) (var foo)))
	case match(lang.AAdd, tPat(lang.TDblUnder, lang.TDblUnder)):
		tup := t.(lang.Tuple)[1].(lang.Tuple)
		v1, err := mod.build(tup[0])
		if err != nil {
			return Val{}, err
		}
		v2, err := mod.build(tup[1])
		if err != nil {
			return Val{}, err
		}
		return Val{v: mod.b.CreateAdd(v1.v, v2.v, "")}, nil

	default:
		return Val{}, fmt.Errorf("unknown compiler command %v", t)
	}
}

// Dump dumps the Module's IR to stdout
func (mod *Module) Dump() {
	mod.m.Dump()
}

// Run executes the Module
// TODO input and output?
func (mod *Module) Run() (interface{}, error) {
	engine, err := llvm.NewExecutionEngine(mod.m)
	if err != nil {
		return nil, err
	}
	defer engine.Dispose()

	funcResult := engine.RunFunction(mod.mainFn, []llvm.GenericValue{})
	defer funcResult.Dispose()
	return funcResult.Int(false), nil
}
