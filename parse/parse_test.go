package parse

import (
	"bytes"
	. "testing"

	"github.com/mediocregopher/ginger/types"
)

func TestReadString(t *T) {
	m := map[string]types.Str{
		`"hey there"`: "hey there",
		`"hey\nthere"`: "hey\nthere",
		`"hey there ⌘"`: "hey there ⌘",
		`"hey\nthere \u2318"`: "hey\nthere ⌘",
	}

	for input, output := range m {
		buf := bytes.NewBufferString(input)
		buf.ReadByte()

		parseOut, err := ReadString(buf)
		if err != nil {
			t.Fatal(err)
		}
		if output != parseOut {
			t.Fatalf("`%s` != `%s`", output, parseOut)
		}
	}
}

func TestParseBareElement(t *T) {
	m := map[string]types.Elem{
		`1`: types.Int(1),
		`12`: types.Int(12),
		`-1`: types.Int(-1),
		`-12`: types.Int(-12),

		`1.0`: types.Float(1.0),
		`12.5`: types.Float(12.5),
		`-12.5`: types.Float(-12.5),

		`-`: types.Str(":-"),

		`bare`: types.Str(":bare"),
		`:not-bare`: types.Str(":not-bare"),
	}

	for input, output := range m {
		el, err := ParseBareElement(input)
		if err != nil {
			t.Fatal(err)
		}
		if output != el {
			t.Fatalf("`%s` != `%s`", output, el)
		}
	}
}
