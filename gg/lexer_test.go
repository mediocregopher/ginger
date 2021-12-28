package gg

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLexer(t *testing.T) {

	expErr := errors.New("eof")

	tests := []struct {
		in  string
		exp []LexerToken
	}{
		{in: "", exp: []LexerToken{}},
		{in: "* fooo\n", exp: []LexerToken{}},
		{
			in: "foo",
			exp: []LexerToken{
				{
					Kind:     LexerTokenKindName,
					Value:    "foo",
					Location: LexerLocation{Row: 0, Col: 0},
				},
			},
		},
		{
			in: "foo bar\nf-o f0O Foo",
			exp: []LexerToken{
				{
					Kind:     LexerTokenKindName,
					Value:    "foo",
					Location: LexerLocation{Row: 0, Col: 0},
				},
				{
					Kind:     LexerTokenKindName,
					Value:    "bar",
					Location: LexerLocation{Row: 0, Col: 4},
				},
				{
					Kind:     LexerTokenKindName,
					Value:    "f-o",
					Location: LexerLocation{Row: 1, Col: 0},
				},
				{
					Kind:     LexerTokenKindName,
					Value:    "f0O",
					Location: LexerLocation{Row: 1, Col: 4},
				},
				{
					Kind:     LexerTokenKindName,
					Value:    "Foo",
					Location: LexerLocation{Row: 1, Col: 8},
				},
			},
		},
		{
			in: "1 100 -100",
			exp: []LexerToken{
				{
					Kind:     LexerTokenKindNumber,
					Value:    "1",
					Location: LexerLocation{Row: 0, Col: 0},
				},
				{
					Kind:     LexerTokenKindNumber,
					Value:    "100",
					Location: LexerLocation{Row: 0, Col: 2},
				},
				{
					Kind:     LexerTokenKindNumber,
					Value:    "-100",
					Location: LexerLocation{Row: 0, Col: 6},
				},
			},
		},
		{
			in: "1<2!-3 ()",
			exp: []LexerToken{
				{
					Kind:     LexerTokenKindNumber,
					Value:    "1",
					Location: LexerLocation{Row: 0, Col: 0},
				},
				{
					Kind:     LexerTokenKindPunctuation,
					Value:    "<",
					Location: LexerLocation{Row: 0, Col: 1},
				},
				{
					Kind:     LexerTokenKindNumber,
					Value:    "2",
					Location: LexerLocation{Row: 0, Col: 2},
				},
				{
					Kind:     LexerTokenKindPunctuation,
					Value:    "!",
					Location: LexerLocation{Row: 0, Col: 3},
				},
				{
					Kind:     LexerTokenKindNumber,
					Value:    "-3",
					Location: LexerLocation{Row: 0, Col: 4},
				},
				{
					Kind:     LexerTokenKindPunctuation,
					Value:    "(",
					Location: LexerLocation{Row: 0, Col: 7},
				},
				{
					Kind:     LexerTokenKindPunctuation,
					Value:    ")",
					Location: LexerLocation{Row: 0, Col: 8},
				},
			},
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {

			lexer := NewLexer(&mockReader{body: []byte(test.in), err: expErr})

			for i := range test.exp {
				tok, err := lexer.Next()
				assert.NoError(t, err)
				assert.Equal(t, test.exp[i], tok, "test.exp[%d]", i)
			}

			tok, err := lexer.Next()
			assert.ErrorIs(t, err, expErr)
			assert.Equal(t, LexerToken{}, tok)

			lexErr := new(LexerError)
			assert.True(t, errors.As(err, &lexErr))

			inParts := strings.Split(test.in, "\n")

			assert.ErrorIs(t, lexErr, expErr)
			assert.Equal(t, lexErr.Location.Row, len(inParts)-1)
			assert.Equal(t, lexErr.Location.Col, len(inParts[len(inParts)-1]))
		})
	}

}
