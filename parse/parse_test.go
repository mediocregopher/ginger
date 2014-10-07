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
