package vm

import (
	"errors"
	"fmt"
	"sync"

	"github.com/mediocregopher/ginger/lang"

	"llvm.org/llvm/bindings/go/llvm"
)

// Types supported by the vm in addition to those which are part of lang
var (
	Atom  = lang.AAtom
	Tuple = lang.ATuple
	Int   = lang.Atom("int")
)

// Ops supported by the vm
var (
	Add    = lang.Atom("add")
	Assign = lang.Atom("assign")
	Var    = lang.Atom("var")
)

////////////////////////////////////////////////////////////////////////////////

// Module contains a compiled set of code which can be run, dumped in IR form,
// or compiled. A Module should be Dispose()'d of once it's no longer being
// used.
type Module struct {
	b      llvm.Builder
	m      llvm.Module
	ctx    varCtx
	mainFn llvm.Value
}

var initOnce sync.Once

// Build creates a new Module by compiling the given Terms as code
// TODO only take in a single Term, implement List and use that with a do op
func Build(tt ...lang.Term) (*Module, error) {
	initOnce.Do(func() {
		llvm.LinkInMCJIT()
		llvm.InitializeNativeTarget()
		llvm.InitializeNativeAsmPrinter()
	})
	mod := &Module{
		b:   llvm.NewBuilder(),
		m:   llvm.NewModule(""),
		ctx: varCtx{},
	}

	var err error
	if mod.mainFn, err = mod.buildFn(tt...); err != nil {
		mod.Dispose()
		return nil, err
	}

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

// TODO make this return a val once we get function types
func (mod *Module) buildFn(tt ...lang.Term) (llvm.Value, error) {
	if len(tt) == 0 {
		return llvm.Value{}, errors.New("function cannot be empty")
	}

	ops := make([]op, len(tt))
	var err error
	for i := range tt {
		if ops[i], err = termToOp(mod.ctx, tt[i]); err != nil {
			return llvm.Value{}, err
		}
	}

	var llvmIns []llvm.Type
	if in := ops[0].inType(); in.llvm.TypeKind() == llvm.VoidTypeKind {
		llvmIns = []llvm.Type{}
	} else {
		llvmIns = []llvm.Type{in.llvm}
	}
	llvmOut := ops[len(ops)-1].outType().llvm

	fn := llvm.AddFunction(mod.m, "", llvm.FunctionType(llvmOut, llvmIns, false))
	block := llvm.AddBasicBlock(fn, "")
	mod.b.SetInsertPoint(block, block.FirstInstruction())

	var out llvm.Value
	for i := range ops {
		if out, err = ops[i].build(mod); err != nil {
			return llvm.Value{}, err
		}
	}
	mod.b.CreateRet(out)
	return fn, nil
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
