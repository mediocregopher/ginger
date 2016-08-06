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

	// setup our context, builder, and module
	bctx := expr.BuildCtx{
		C: expr.NewCtx(),
		B: llvm.NewBuilder(),
		M: llvm.NewModule("my_module"),
	}

	// do the work in the function
	tup := expr.NewTuple(expr.Int(1), expr.Int(2))
	addMacro := expr.Macro("add")
	stmt := expr.Statement{Op: addMacro, Arg: tup}
	stmt = expr.Statement{Op: addMacro, Arg: expr.NewTuple(stmt, expr.Int(3))}

	//block := expr.Block([]expr.Expr{stmt})
	//fn := block.LLVMVal(expr.RootCtx, lctx)

	// create main and call our function
	mainFn := llvm.AddFunction(bctx.M, "main", llvm.FunctionType(llvm.Int64Type(), []llvm.Type{}, false))
	mainBlock := llvm.AddBasicBlock(mainFn, "entry")
	bctx.B.SetInsertPoint(mainBlock, mainBlock.FirstInstruction())
	v := bctx.Build(stmt)
	bctx.B.CreateRet(v)

	//ret := lctx.B.CreateCall(fn, []llvm.Value{}, "")
	//lctx.B.CreateRet(ret)

	// verify it's all good
	if err := llvm.VerifyModule(bctx.M, llvm.ReturnStatusAction); err != nil {
		panic(err)
	}
	fmt.Println("# verified")

	// Dump the IR
	fmt.Println("# dumping IR")
	bctx.M.Dump()
	fmt.Println("# done dumping IR")

	// create our exe engine
	fmt.Println("# creating new execution engine")
	engine, err := llvm.NewExecutionEngine(bctx.M)
	if err != nil {
		panic(err)
	}

	// run the function!
	fmt.Println("# running the function main")
	funcResult := engine.RunFunction(bctx.M.NamedFunction("main"), []llvm.GenericValue{})
	fmt.Printf("%d\n", funcResult.Int(false))
}
