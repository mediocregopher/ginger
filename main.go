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
	lctx := expr.LLVMCtx{
		B: llvm.NewBuilder(),
		M: llvm.NewModule("my_module"),
	}

	// do the work in the function
	a := expr.Expr{Actual: expr.Int(1)}
	b := expr.Expr{Actual: expr.Int(2)}
	c := expr.Expr{Actual: expr.Int(3)}
	tup := expr.Expr{Actual: expr.Tuple([]expr.Expr{a, b, c})}
	addMacro := expr.Expr{Actual: expr.Macro("add")}
	stmt := expr.Expr{Actual: expr.Statement{Op: addMacro, Arg: tup}}

	//block := expr.Block([]expr.Expr{stmt})
	//fn := block.LLVMVal(expr.RootCtx, lctx)

	// create main and call our function
	mainFn := llvm.AddFunction(lctx.M, "main", llvm.FunctionType(llvm.Int64Type(), []llvm.Type{}, false))
	mainBlock := llvm.AddBasicBlock(mainFn, "entry")
	lctx.B.SetInsertPoint(mainBlock, mainBlock.FirstInstruction())
	expr.BuildStmt(expr.RootCtx, lctx, stmt)
	lctx.B.CreateRet(expr.RootCtx.LastVal)

	//ret := lctx.B.CreateCall(fn, []llvm.Value{}, "")
	//lctx.B.CreateRet(ret)

	// verify it's all good
	if err := llvm.VerifyModule(lctx.M, llvm.ReturnStatusAction); err != nil {
		panic(err)
	}
	fmt.Println("# verified")

	// Dump the IR
	fmt.Println("# dumping IR")
	lctx.M.Dump()
	fmt.Println("# done dumping IR")

	// create our exe engine
	fmt.Println("# creating new execution engine")
	engine, err := llvm.NewExecutionEngine(lctx.M)
	if err != nil {
		panic(err)
	}

	// run the function!
	fmt.Println("# running the function main")
	funcResult := engine.RunFunction(lctx.M.NamedFunction("main"), []llvm.GenericValue{})
	fmt.Printf("%d\n", funcResult.Int(false))
}
