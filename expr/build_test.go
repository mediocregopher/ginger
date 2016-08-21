package expr

import (
	"fmt"
	. "testing"

	"llvm.org/llvm/bindings/go/llvm"
)

func buildTest(t *T, expected int64, stmts ...Statement) {
	fmt.Println("-----------------------------------------")
	ctx := NewCtx()
	bctx := NewBuildCtx("")

	fn := llvm.AddFunction(bctx.M, "", llvm.FunctionType(llvm.Int64Type(), []llvm.Type{}, false))
	fnbl := llvm.AddBasicBlock(fn, "")
	bctx.B.SetInsertPoint(fnbl, fnbl.FirstInstruction())
	out := bctx.Build(ctx, stmts...)
	bctx.B.CreateRet(out)

	fmt.Println("######## dumping IR")
	bctx.M.Dump()
	fmt.Println("######## done dumping IR")

	if err := llvm.VerifyModule(bctx.M, llvm.ReturnStatusAction); err != nil {
		t.Fatal(err)
	}

	eng, err := llvm.NewExecutionEngine(bctx.M)
	if err != nil {
		t.Fatal(err)
	}

	res := eng.RunFunction(fn, []llvm.GenericValue{}).Int(false)
	if int64(res) != expected {
		t.Errorf("expected:[%T]%v actual:[%T]%v", expected, expected, res, res)
	}
}

func TestAdd(t *T) {
	buildTest(t, 2,
		NewStatement(Macro("add"), Int(1), Int(1)))
	buildTest(t, 4,
		NewStatement(Macro("add"), Int(1),
			NewStatement(Macro("add"), Int(1), Int(2))))
	buildTest(t, 6,
		NewStatement(Macro("add"),
			NewStatement(Macro("add"), Int(1), Int(2)),
			NewStatement(Macro("add"), Int(1), Int(2))))
}

func TestBind(t *T) {
	buildTest(t, 2,
		NewStatement(Macro("bind"), Identifier("A"), Int(1)),
		NewStatement(Macro("add"), Identifier("A"), Int(1)))
	buildTest(t, 2,
		NewStatement(Macro("bind"), Identifier("A"), Int(1)),
		NewStatement(Macro("add"), Identifier("A"), Identifier("A")))
	buildTest(t, 2,
		NewStatement(Macro("bind"), Identifier("A"), NewTuple(Int(1), Int(1))),
		NewStatement(Macro("add"), Identifier("A")))
	buildTest(t, 3,
		NewStatement(Macro("bind"), Identifier("A"), NewTuple(Int(1), Int(1))),
		NewStatement(Macro("add"), Int(1),
			NewStatement(Macro("add"), Identifier("A"))))
	buildTest(t, 4,
		NewStatement(Macro("bind"), Identifier("A"), NewTuple(Int(1), Int(1))),
		NewStatement(Macro("add"),
			NewStatement(Macro("add"), Identifier("A")),
			NewStatement(Macro("add"), Identifier("A"))))
}

func TestOp(t *T) {
	incr := NewStatement(Macro("op"),
		NewList(
			NewStatement(Macro("add"), Int(1), NewStatement(Macro("in"))),
		),
	)

	// bound op
	buildTest(t, 2,
		NewStatement(Macro("bind"), Identifier("incr"), incr),
		NewStatement(Identifier("incr"), Int(1)))

	// double bound op
	buildTest(t, 3,
		NewStatement(Macro("bind"), Identifier("incr"), incr),
		NewStatement(Identifier("incr"),
			NewStatement(Identifier("incr"), Int(1))))

	// anon op
	buildTest(t, 2,
		NewStatement(incr, Int(1)))

	// double anon op
	buildTest(t, 3,
		NewStatement(incr,
			NewStatement(incr, Int(1))))
}
