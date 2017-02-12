package main

import (
	"fmt"

	"github.com/mediocregopher/ginger/lang"
	"github.com/mediocregopher/ginger/vm"
)

func main() {
	mkcmd := func(a lang.Atom, args ...lang.Term) lang.Tuple {
		return lang.Tuple{a, lang.Tuple{vm.Tuple, lang.Tuple(args)}}
	}
	mkint := func(i string) lang.Tuple {
		return lang.Tuple{vm.Int, lang.Const(i)}
	}

	t := mkcmd(vm.Add, mkint("1"),
		mkcmd(vm.Add, mkint("2"), mkint("3")))

	mod, err := vm.Build(t)
	if err != nil {
		panic(err)
	}
	defer mod.Dispose()

	mod.Dump()

	out, err := mod.Run()
	fmt.Printf("\n\n########\nout: %v %v\n", out, err)
}
