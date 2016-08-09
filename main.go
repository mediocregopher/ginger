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
	ctx := expr.NewCtx()
	bctx := expr.BuildCtx{
		B: llvm.NewBuilder(),
		M: llvm.NewModule("my_module"),
	}

	// do the work in the function
	add := expr.Macro("add")
	bind := expr.Macro("bind")
	do := expr.Macro("do")
	ctxnew := expr.Macro("ctxnew")
	ctxbind := expr.Macro("ctxbind")
	ctxget := expr.Macro("ctxget")

	ctx1 := expr.Identifier("ctx1")
	ctx2 := expr.Identifier("ctx2")
	idA := expr.Identifier("A")
	idB := expr.Identifier("B")

	//myAdd := expr.Identifier("myAdd")
	out := expr.Identifier("out")
	// TODO we couldn't actually use this either, because the builder was
	// changing out the internal values of the List the first time it was hit,
	// and then just using those the second time around
	//myAddStmts := expr.NewList(
	//	expr.NewStatement(bind, out, expr.NewStatement(add, idA, idB)),
	//)

	stmts := []expr.Statement{
		// TODO revisit how bind and related macros (maybe all macros?) deal
		// with arguments and their evaluation (keeping an identifier vs
		// eval-ing it)
		//expr.NewStatement(bind, myAdd, myAddStmts),

		expr.NewStatement(bind, ctx1, expr.NewStatement(ctxnew)),
		expr.NewStatement(ctxbind, ctx1, idA, expr.Int(1)),
		expr.NewStatement(ctxbind, ctx1, idB, expr.Int(2)),
		expr.NewStatement(do, ctx1, expr.NewList(
			expr.NewStatement(bind, out, expr.NewStatement(add, idA, idB)),
		)),

		expr.NewStatement(bind, ctx2, expr.NewStatement(ctxnew)),
		expr.NewStatement(ctxbind, ctx2, idA, expr.Int(3)),
		expr.NewStatement(ctxbind, ctx2, idB, expr.Int(4)),
		expr.NewStatement(do, ctx2, expr.NewList(
			expr.NewStatement(bind, out, expr.NewStatement(add, idA, idB)),
		)),

		expr.NewStatement(
			add,
			expr.NewStatement(ctxget, ctx1, out),
			expr.NewStatement(ctxget, ctx2, out),
		),
	}

	// create main and call our function
	mainFn := llvm.AddFunction(bctx.M, "main", llvm.FunctionType(llvm.Int64Type(), []llvm.Type{}, false))
	mainBlock := llvm.AddBasicBlock(mainFn, "entry")
	bctx.B.SetInsertPoint(mainBlock, mainBlock.FirstInstruction())
	v := bctx.Build(ctx, stmts...)
	bctx.B.CreateRet(v)

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
