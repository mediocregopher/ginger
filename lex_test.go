package ginger

import (
	"bytes"
	"io"
	. "testing"

	"github.com/mediocregopher/lexgo"
	"github.com/stretchr/testify/assert"
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

	/* block comment */
	prefix /*
		Another block comment
		/* Embedded */
		/*
			Super embedded
		*/
	*/ suffix

	// this one is kind of fun, technically it's a comment
	/*/

	(punctuation,is{cool}<> )
	-tab
`

func TestLex(t *T) {
	l := newLexer(bytes.NewBufferString(lexTestSrc))

	assertNext := func(typ lexgo.TokenType, val string) {
		t.Logf("asserting %q", val)
		tok := l.Next()
		assert.Equal(t, typ, tok.TokenType)
		assert.Equal(t, val, tok.Val)
	}

	assertNext(identifier, "a")
	assertNext(identifier, "anIdentifier")
	assertNext(number, "1")
	assertNext(number, "100")
	assertNext(number, "1.5")
	assertNext(number, "1.5e9")
	assertNext(identifier, "prefix")
	assertNext(identifier, "suffix")
	assertNext(punctuation, "(")
	assertNext(identifier, "punctuation")
	assertNext(punctuation, ",")
	assertNext(identifier, "is")
	assertNext(punctuation, "{")
	assertNext(identifier, "cool")
	assertNext(punctuation, "}")
	assertNext(punctuation, "<")
	assertNext(punctuation, ">")
	assertNext(punctuation, ")")
	assertNext(identifier, "-tab")

	tok := l.Next()
	assert.Equal(t, tok.TokenType, lexgo.Err)
	assert.Equal(t, tok.Err, io.EOF)
}
