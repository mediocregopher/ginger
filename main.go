package main

import (
	"fmt"

	"github.com/mediocregopher/ginger/expr"

	"llvm.org/llvm/bindings/go/llvm"
)

type addActual []expr.Expr

func (aa addActual) Equal(expr.Actual) bool { return false }

func (aa addActual) LLVMVal(builder llvm.Builder) llvm.Value {
	a := builder.CreateLoad(aa[0].LLVMVal(builder), "")
	for i := range aa[1:] {
		b := builder.CreateLoad(aa[i+1].LLVMVal(builder), "")
		a = builder.CreateAdd(a, b, "")
	}
	return a
}

func main() {
	//ee, err := expr.Parse(os.Stdin)
	//if err != nil {
	//	panic(err)
	//}
	//for _, e := range ee {
	//	fmt.Println(e)
	//}

	llvm.LinkInMCJIT()
	llvm.InitializeNativeTarget()
	llvm.InitializeNativeAsmPrinter()

	// setup our builder and module
	builder := llvm.NewBuilder()
	mod := llvm.NewModule("my_module")

	// create our function prologue
	main := llvm.FunctionType(llvm.Int64Type(), []llvm.Type{}, false)
	llvm.AddFunction(mod, "main", main)
	block := llvm.AddBasicBlock(mod.NamedFunction("main"), "entry")
	builder.SetInsertPoint(block, block.FirstInstruction())

	a := expr.Expr{Actual: expr.Int(1)}
	b := expr.Expr{Actual: expr.Int(2)}
	c := expr.Expr{Actual: expr.Int(3)}
	add := addActual{a, b, c}
	result := add.LLVMVal(builder)
	builder.CreateRet(result)

	// verify it's all good
	if err := llvm.VerifyModule(mod, llvm.ReturnStatusAction); err != nil {
		panic(err)
	}
	fmt.Println("# verified")

	// Dump the IR
	fmt.Println("# dumping IR")
	mod.Dump()
	fmt.Println("# done dumping IR")

	// create our exe engine
	fmt.Println("# creating new execution engine")
	engine, err := llvm.NewExecutionEngine(mod)
	if err != nil {
		panic(err)
	}

	// run the function!
	fmt.Println("# running the function main")
	funcResult := engine.RunFunction(mod.NamedFunction("main"), []llvm.GenericValue{})
	fmt.Printf("%d\n", funcResult.Int(false))
}
