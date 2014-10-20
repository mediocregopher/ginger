package parse

import (
	"bytes"
	"io"
	. "testing"

	"github.com/mediocregopher/ginger/seq"
	"github.com/mediocregopher/ginger/types"
)

func TestParseBareString(t *T) {
	m := map[string][]types.Elem{
		"foo": []types.Elem{types.GoType{":foo"}},

		"foo bar": []types.Elem{
			types.GoType{":foo"},
			types.GoType{":bar"},
		},

		"foo \"bar\"": []types.Elem{
			types.GoType{":foo"},
			types.GoType{"bar"},
		},

		"()": []types.Elem{seq.NewList()},

		"(foo)": []types.Elem{seq.NewList(
			types.GoType{":foo"},
		)},

		"(foo (bar))": []types.Elem{seq.NewList(
			types.GoType{":foo"},
			seq.NewList(types.GoType{":bar"}),
		)},

		"{}": []types.Elem{seq.NewHashMap()},

		"{foo bar}": []types.Elem{seq.NewHashMap(
			seq.KeyVal(types.GoType{":foo"}, types.GoType{":bar"}),
		)},
	}

	for input, output := range m {
		buf := bytes.NewBufferString(input)
		p := NewParser(buf)
		parsed := make([]types.Elem, 0, len(output))
		for {
			el, err := p.ReadElem()
			if err == io.EOF {
				break
			} else if err != nil {
				t.Fatal(err)
			}
			parsed = append(parsed, el)
		}

		if len(output) != len(parsed) {
			t.Fatalf("input: %q %#v != %#v", input, output, parsed)
		}

		for i := range output {
			if !output[i].Equal(parsed[i]) {
				t.Fatalf("input: %q (%d) %#v != %#v", input, i, output[i], parsed[i])
			}
		}
	}
}
