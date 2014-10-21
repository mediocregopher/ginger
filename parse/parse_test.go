package parse

import (
	"bytes"
	"io"
	. "testing"

	"github.com/mediocregopher/ginger/seq"
	"github.com/mediocregopher/ginger/types"
)

func TestParse(t *T) {
	m := map[string]types.Elem{
		"1":  types.GoType{int(1)},
		"-1": types.GoType{int(-1)},
		"+1": types.GoType{int(1)},

		"1.5":   types.GoType{float32(1.5)},
		"-1.5":  types.GoType{float32(-1.5)},
		"+1.5":  types.GoType{float32(1.5)},
		"1.5e1": types.GoType{float32(15)},

		"foo": types.GoType{":foo"},

		"()": seq.NewList(),

		"(foo)": seq.NewList(
			types.GoType{":foo"},
		),

		"(foo (bar))": seq.NewList(
			types.GoType{":foo"},
			seq.NewList(types.GoType{":bar"}),
		),

		"{}": seq.NewHashMap(),

		"{foo bar}": seq.NewHashMap(
			seq.KeyVal(types.GoType{":foo"}, types.GoType{":bar"}),
		),
	}

	for input, output := range m {
		parsed, err := ParseString(input)
		if err != nil {
			t.Fatal(err)
		}

		if !output.Equal(parsed) {
			t.Fatalf("input: %q %#v != %#v", input, output, parsed)
		}
	}
}

func TestParseMulti(t *T) {
	m := map[string][]types.Elem{
		"foo 4 bar": {
			types.GoType{":foo"},
			types.GoType{4},
			types.GoType{":bar"},
		},

		"foo \"bar\"": {
			types.GoType{":foo"},
			types.GoType{"bar"},
		},
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
