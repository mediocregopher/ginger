package eval

import (
	. "testing"

	"github.com/mediocregopher/ginger/macros/pkgctx"
	"github.com/mediocregopher/ginger/parse"
	"github.com/mediocregopher/ginger/seq"
	"github.com/mediocregopher/ginger/types"
)

// This is NOT how I want eval to really work in the end, but I wanted to get
// something down before I kept thinking about it, so I would know what would
// work

func shittyPlus(s seq.Seq) types.Elem {
	fn := func(acc, el types.Elem) (types.Elem, bool) {
		i := acc.(types.GoType).V.(int) + el.(types.GoType).V.(int)
		return types.GoType{i}, false
	}

	return seq.Reduce(fn, types.GoType{0}, s)
}

func TestShittyPlus(t *T) {
	p := &pkgctx.PkgCtx{
		CallMap: map[string]interface{}{
			"shittyPlus": Evaler(shittyPlus),
		},
	}

	m := map[string]types.Elem{
		"(: shittyPlus)":       types.GoType{0},
		"(: shittyPlus 1 2 3)": types.GoType{6},
		`(: shittyPlus 1 2 3
			(: shittyPlus 1 2 3))`: types.GoType{12},
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
