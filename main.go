package main

import (
	"fmt"

	"github.com/mediocregopher/ginger/lang"
	"github.com/mediocregopher/ginger/vm"
)

func main() {
	t := lang.Tuple{vm.Add, lang.Tuple{vm.Tuple, lang.Tuple{
		lang.Tuple{vm.Int, lang.Const("1")},
		lang.Tuple{vm.Int, lang.Const("2")},
	}}}

	mod, err := vm.Build(t)
	if err != nil {
		panic(err)
	}
	defer mod.Dispose()

	mod.Dump()

	out, err := mod.Run()
	fmt.Printf("\n\n########\nout: %v %v\n", out, err)
}
