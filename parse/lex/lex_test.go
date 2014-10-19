package lex

import (
	"bytes"
	. "testing"
)

func TestLexer(t *T) {
	m := map[string][]Token{
		"": {{eof, ""}},
		" \t": {{eof, ""}},
		"a b c": {{BareString, "a"},
				  {BareString, "b"},
				  {BareString, "c"},
				  {eof, ""}},
		"\"foo\" bar": {{QuotedString, "\"foo\""},
						{BareString, "bar"},
						{eof, ""}},
		"\"foo\nbar\" baz": {{QuotedString, "\"foo\nbar\""},
							 {BareString, "baz"},
							 {eof, ""}},
		"( foo bar ) baz": {{Open, "("},
						    {BareString, "foo"},
						    {BareString, "bar"},
						    {Close, ")"},
						    {BareString, "baz"},
						    {eof, ""}},
		"((foo-bar))":     {{Open, "("},
							{Open, "("},
							{BareString, "foo-bar"},
							{Close, ")"},
							{Close, ")"},
						    {eof, ""}},
		"(\"foo\nbar\")":  {{Open, "("},
							{QuotedString, "\"foo\nbar\""},
							{Close, ")"},
							{eof, ""}},
	}

	for input, output := range m {
		buf := bytes.NewBufferString(input)
		l := NewLexer(buf)
		for i := range output {
			tok := l.Next()
			if tok == nil {
				if output[i].Type == eof {
					continue
				}
				t.Fatalf("input: %q (%d) %#v != %#v", input, i, output[i], tok)
			}
			if *tok != output[i] {
				t.Fatalf("input: %s (%d) %#v != %#v", input, i, output[i], tok)
			}
		}
	}
}