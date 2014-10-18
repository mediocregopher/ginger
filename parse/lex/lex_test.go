package lexer

import (
	"bytes"
	. "testing"
)

func TestLexer(t *T) {
	m := map[string][]Token{
		"": {{EOF, ""}},
		" \t": {{EOF, ""}},
		"a b c": {{BareString, "a"},
				  {BareString, "b"},
				  {BareString, "c"},
				  {EOF, ""}},
		"\"foo\" bar": {{QuotedString, "\"foo\""},
						{BareString, "bar"},
						{EOF, ""}},
		"\"foo\nbar\" baz": {{QuotedString, "\"foo\nbar\""},
							 {BareString, "baz"},
							 {EOF, ""}},
		"( foo bar ) baz": {{OpenParen, "("},
						    {BareString, "foo"},
						    {BareString, "bar"},
						    {CloseParen, ")"},
						    {BareString, "baz"},
						    {EOF, ""}},
		"((foo-bar))":     {{OpenParen, "("},
							{OpenParen, "("},
							{BareString, "foo-bar"},
							{CloseParen, ")"},
							{CloseParen, ")"},
						    {EOF, ""}},
		"(\"foo\nbar\")":  {{OpenParen, "("},
							{QuotedString, "\"foo\nbar\""},
							{CloseParen, ")"},
							{EOF, ""}},
	}

	for input, output := range m {
		buf := bytes.NewBufferString(input)
		l := NewLexer(buf)
		for i := range output {
			tok := l.Next()
			if tok == nil {
				if output[i].Type == EOF {
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
