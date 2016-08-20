package main

import (
	"fmt"
	"log"

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

	log.Printf("initializing llvm")
	llvm.LinkInMCJIT()
	llvm.InitializeNativeTarget()
	llvm.InitializeNativeAsmPrinter()

	log.Printf("initializing build context")
	ctx := expr.NewCtx()
	bctx := expr.BuildCtx{
		B: llvm.NewBuilder(),
		M: llvm.NewModule("my_module"),
	}

	log.Printf("making program")
	add := expr.Macro("add")
	bind := expr.Macro("bind")
	op := expr.Macro("op")
	in := expr.Macro("in")
	incr := expr.Identifier("incr")
	stmts := []expr.Statement{
		expr.NewStatement(bind, incr,
			expr.NewStatement(op,
				expr.NewList(
					expr.NewStatement(add, expr.NewTuple(
						expr.Int(1),
						expr.NewStatement(in, expr.NewTuple()),
					)),
				),
			),
		),

		expr.NewStatement(
			incr,
			expr.Int(5),
		),
	}

	log.Printf("creating main function")
	mainFn := llvm.AddFunction(bctx.M, "main", llvm.FunctionType(llvm.Int64Type(), []llvm.Type{}, false))
	mainBlock := llvm.AddBasicBlock(mainFn, "entry")
	bctx.B.SetInsertPoint(mainBlock, mainBlock.FirstInstruction())
	log.Printf("actually processing program")
	out := bctx.Build(ctx, stmts...)
	bctx.B.CreateRet(out)
	//bctx.Build(ctx, stmts...)
	//bctx.B.CreateRet(llvm.ConstInt(llvm.Int64Type(), uint64(5), false))

	fmt.Println("######## dumping IR")
	bctx.M.Dump()
	fmt.Println("######## done dumping IR")

	log.Printf("verifying")
	if err := llvm.VerifyModule(bctx.M, llvm.ReturnStatusAction); err != nil {
		panic(err)
	}

	log.Printf("creating execution enging")
	engine, err := llvm.NewExecutionEngine(bctx.M)
	if err != nil {
		panic(err)
	}

	log.Printf("running main function")
	funcResult := engine.RunFunction(bctx.M.NamedFunction("main"), []llvm.GenericValue{})
	fmt.Printf("\nOUTPUT:\n%d\n", funcResult.Int(false))
}
