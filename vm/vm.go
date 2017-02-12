package vm

import (
	"errors"
	"fmt"
	"sync"

	"github.com/mediocregopher/ginger/lang"

	"llvm.org/llvm/bindings/go/llvm"
)

type val struct {
	typ lang.Term
	v   llvm.Value
}

// Module contains a compiled set of code which can be run, dumped in IR form,
// or compiled. A Module should be Dispose()'d of once it's no longer being
// used.
type Module struct {
	b         llvm.Builder
	m         llvm.Module
	mainFn    llvm.Value
	buildCmds []buildCmd
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
	mod.buildCmds = buildCmds(mod)

	var err error
	if mod.mainFn, err = mod.buildFn(t); err != nil {
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

func (mod *Module) matchingBuildCmd(t lang.Term) (buildCmd, error) {
	for _, cmd := range mod.buildCmds {
		if !cmd.matches(t) {
			continue
		}
		return cmd, nil
	}
	return buildCmd{}, fmt.Errorf("unknown compiler command %v", t)
}

func (mod *Module) inType(t lang.Term) (llvm.Type, error) {
	cmd, err := mod.matchingBuildCmd(t)
	if err != nil {
		return llvm.Type{}, err
	}
	return cmd.inType(t.(lang.Tuple)[1])
}

func (mod *Module) outType(t lang.Term) (llvm.Type, error) {
	cmd, err := mod.matchingBuildCmd(t)
	if err != nil {
		return llvm.Type{}, err
	}
	return cmd.outType(t.(lang.Tuple)[1])
}

func (mod *Module) build(t lang.Term) (val, error) {
	cmd, err := mod.matchingBuildCmd(t)
	if err != nil {
		return val{}, err
	}
	return cmd.build(t.(lang.Tuple)[1])
}

// TODO make this return a val once we get function types
func (mod *Module) buildFn(tt ...lang.Term) (llvm.Value, error) {
	if len(tt) == 0 {
		return llvm.Value{}, errors.New("function cannot be empty")
	}

	inType, err := mod.inType(tt[0])
	if err != nil {
		return llvm.Value{}, err
	}
	var inTypes []llvm.Type
	if inType.TypeKind() == llvm.VoidTypeKind {
		inTypes = []llvm.Type{}
	} else {
		inTypes = []llvm.Type{inType}
	}

	outType, err := mod.outType(tt[len(tt)-1])
	if err != nil {
		return llvm.Value{}, err
	}

	fn := llvm.AddFunction(mod.m, "", llvm.FunctionType(outType, inTypes, false))
	block := llvm.AddBasicBlock(fn, "")
	mod.b.SetInsertPoint(block, block.FirstInstruction())

	var out val
	for _, t := range tt {
		if out, err = mod.build(t); err != nil {
			return llvm.Value{}, err
		}
	}
	mod.b.CreateRet(out.v)
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
