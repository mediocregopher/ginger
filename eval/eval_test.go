package eval

import (
	. "testing"

	"github.com/mediocregopher/ginger/core"
	"github.com/mediocregopher/ginger/macros/pkgctx"
	"github.com/mediocregopher/ginger/parse"
	"github.com/mediocregopher/ginger/types"
)

// This is NOT how I want eval to really work in the end, but I wanted to get
// something down before I kept thinking about it, so I would know what would
// work

func TestShittyPlus(t *T) {
	p := &pkgctx.PkgCtx{
		CallMap: map[string]interface{}{
			"Plus": Evaler(core.Plus),
		},
	}

	m := map[string]types.Elem{
		"(: Plus)":       types.GoType{0},
		"(: Plus 1 2 3)": types.GoType{6},
		`(: Plus 1 2 3
			(: Plus 1 2 3))`: types.GoType{12},
	}

	for input, output := range m {
		parsed, err := parse.ParseString(input)
		if err != nil {
			t.Fatal(err)
		}

		evald := Eval(p, parsed)
		if !evald.Equal(output) {
			t.Fatalf("input: %q %#v != %#v", input, output, evald)
		}
	}
}
