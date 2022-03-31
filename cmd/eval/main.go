package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/mediocregopher/ginger/gg"
	"github.com/mediocregopher/ginger/vm"
)

func main() {

	if len(os.Args) < 3 {
		fmt.Printf(`Usage: %s <operation source> "in = <value>"\n`, os.Args[0])
		return
	}

	opSrc := os.Args[1]
	inSrc := os.Args[2]

	inVal, err := gg.DecodeSingleValueFromLexer(
		gg.NewLexer(bytes.NewBufferString(inSrc + ";")),
	)

	if err != nil {
		panic(fmt.Sprintf("decoding input: %v", err))
	}

	res, err := vm.EvaluateSource(
		bytes.NewBufferString(opSrc),
		vm.Value{Value: inVal},
		vm.GlobalScope,
	)

	if err != nil {
		panic(fmt.Sprintf("evaluating: %v", err))
	}

	fmt.Println(res)
}
