package main

import (
	"fmt"

	"github.com/mediocregopher/ginger/expr"

	"llvm.org/llvm/bindings/go/llvm"
)

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
	tup := expr.Expr{Actual: expr.Tuple{a, b, c}}
	addMacro := expr.Expr{Actual: expr.Macro("add")}
	stmt := expr.Expr{Actual: expr.Statement{In: tup, To: addMacro}}

	result := stmt.LLVMVal(expr.RootCtx, builder)
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
