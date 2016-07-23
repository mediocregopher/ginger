package lexer

import (
	"bytes"
	. "testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var lexTestSrc = `
	// this is a comment
	// // this is also a comment
	a
	anIdentifier
	1
	100
	1.5
	1.5e9

	/*
		some stuff
	*/

	/* this should actually work */
	/*/

	/*
		nested!
		/*
			wtf this is crazy
		*/
	*/

	(punctuation,is{cool}<> )
	-tab

	"this is a string", "and so is this one"
	"\"foo"
	"bar\"baz\""
	"buz\0"
`

func TestLex(t *T) {
	l := New(bytes.NewBufferString(lexTestSrc))

	assertNext := func(typ TokenType, val string, row, col int) {
		t.Logf("asserting %s %q [row:%d col:%d]", typ, val, row, col)
		require.True(t, l.HasNext())
		tok := l.Next()
		assert.Equal(t, typ, tok.TokenType)
		assert.Equal(t, val, tok.Val)
		assert.Equal(t, row, tok.Row)
		assert.Equal(t, col, tok.Col)
	}

	assertNext(Identifier, "a", 3, 2)
	assertNext(Identifier, "anIdentifier", 4, 2)
	assertNext(Identifier, "1", 5, 2)
	assertNext(Identifier, "100", 6, 2)
	assertNext(Identifier, "1.5", 7, 2)
	assertNext(Identifier, "1.5e9", 8, 2)
	assertNext(Punctuation, "(", 24, 2)
	assertNext(Identifier, "punctuation", 24, 3)
	assertNext(Punctuation, ",", 24, 14)
	assertNext(Identifier, "is", 24, 15)
	assertNext(Punctuation, "{", 24, 17)
	assertNext(Identifier, "cool", 24, 18)
	assertNext(Punctuation, "}", 24, 22)
	assertNext(Punctuation, "<", 24, 23)
	assertNext(Punctuation, ">", 24, 24)
	assertNext(Punctuation, ")", 24, 26)
	assertNext(Identifier, "-tab", 25, 2)
	assertNext(String, `"this is a string"`, 27, 2)
	assertNext(Punctuation, ",", 27, 20)
	assertNext(String, `"and so is this one"`, 27, 22)
	assertNext(String, `"\"foo"`, 28, 2)
	assertNext(String, `"bar\"baz\""`, 29, 2)
	assertNext(String, `"buz\0"`, 30, 2)
	assertNext(EOF, "EOF", 31, 0)

	assert.False(t, l.HasNext())
}
