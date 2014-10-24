package main

import (
	"os"
	"text/template"
)

// Represents a type (like int or float32) which can have a math operation
// called upon it (like + or -)
type MathOpType struct {
	CastFn string
	Type   string
}

// The standard types for which math operations work upon. These don't include
// byte and rune because those are aliases of int8 and int32, respectively
var MathOpTypes = []MathOpType{
	{"Int", "int"},
	{"Int8", "int8"},
	{"Int16", "int16"},
	{"Int32", "int32"},
	{"Int64", "int64"},
	{"Uint", "uint"},
	{"Uint8", "uint8"},
	{"Uint16", "uint16"},
	{"Uint32", "uint32"},
	{"Uint64", "uint64"},
	{"Float32", "float32"},
	{"Float64", "float64"},
	{"Complex64", "complex64"},
	{"Complex128", "complex128"},
}

var MathOpsIntOnly = []MathOpType{
	{"Int", "int"},
	{"Int8", "int8"},
	{"Int16", "int16"},
	{"Int32", "int32"},
	{"Int64", "int64"},
	{"Uint", "uint"},
	{"Uint8", "uint8"},
	{"Uint16", "uint16"},
	{"Uint32", "uint32"},
	{"Uint64", "uint64"},
}

// Represents a single math operation which can be performed (like + or -)
type MathOp struct {
	Public  string
	Private string
	Op      string

	// Will be the first item in the reduce, and allows for the function being
	// called with an empty seq. If empty string than the first item in the
	// given sequence is used and an empty sequence is not allowed
	Unit string

	// This is going to be the same for all ops, it's just convenient to have
	// here
	OpTypes []MathOpType

	// This only applies for plus, which allows for adding two strings together
	IncludeString bool
}

var MathOps = []MathOp{
	{"Plus", "plus", "+", "0", MathOpTypes, true},
	{"Minus", "minus", "-", "", MathOpTypes, false},

	{"Mult", "mult", "*", "1", MathOpTypes, false},
	{"Div", "div", "/", "", MathOpTypes, false},

	{"Mod", "mod", "%", "", MathOpsIntOnly, false},
}

func main() {
	tpl, err := template.ParseFiles("mathgen/mathgen.tpl")
	if err != nil {
		panic(err)
	}
	if err := tpl.Execute(os.Stdout, MathOps); err != nil {
		panic(err)
	}
}
