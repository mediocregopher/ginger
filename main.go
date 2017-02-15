package main

import (
	"fmt"

	"github.com/mediocregopher/ginger/lang"
	"github.com/mediocregopher/ginger/vm"
)

func main() {
	mkcmd := func(a lang.Atom, args ...lang.Term) lang.Tuple {
		if len(args) == 1 {
			return lang.Tuple{a, args[0]}
		}
		return lang.Tuple{a, lang.Tuple(args)}
	}
	mkint := func(i string) lang.Tuple {
		return lang.Tuple{vm.Int, lang.Const(i)}
	}

	//foo := lang.Atom("foo")
	//tt := []lang.Term{
	//	mkcmd(vm.Assign, foo, mkint("1")),
	//	mkcmd(vm.Add, mkcmd(vm.Tuple, mkcmd(vm.Var, foo), mkint("2"))),
	//}

	foo := lang.Atom("foo")
	bar := lang.Atom("bar")
	baz := lang.Atom("baz")
	tt := []lang.Term{
		mkcmd(vm.Assign, foo, mkcmd(vm.Tuple, mkint("1"), mkint("2"))),
		mkcmd(vm.Assign, bar, mkcmd(vm.Add, mkcmd(vm.Var, foo))),
		mkcmd(vm.Assign, baz, mkcmd(vm.Add, mkcmd(vm.Var, foo))),
		mkcmd(vm.Add, mkcmd(vm.Tuple, mkcmd(vm.Var, bar), mkcmd(vm.Var, baz))),
	}

	mod, err := vm.Build(tt...)
	if err != nil {
		panic(err)
	}
	defer mod.Dispose()

	mod.Dump()

	out, err := mod.Run()
	fmt.Printf("\n\n########\nout: %v %v\n", out, err)
}
