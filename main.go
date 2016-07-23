package main

import (
	"fmt"
	"os"

	"github.com/mediocregopher/ginger/expr"
)

func main() {
	ee, err := expr.Parse(os.Stdin)
	if err != nil {
		panic(err)
	}
	for _, e := range ee {
		fmt.Println(e)
	}
}
