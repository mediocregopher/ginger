package ginger

import (
	. "testing"

	"github.com/mediocregopher/ginger/lexer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSliceEnclosedToks(t *T) {
	doAssert := func(in, expOut, expRem []lexer.Token) {
		out, rem, err := sliceEnclosedToks(in, openParen, closeParen)
		require.Nil(t, err)
		assert.Equal(t, expOut, out)
		assert.Equal(t, expRem, rem)
	}
	foo := lexer.Token{TokenType: lexer.Identifier, Val: "foo"}
	bar := lexer.Token{TokenType: lexer.Identifier, Val: "bar"}

	toks := []lexer.Token{openParen, closeParen}
	doAssert(toks, []lexer.Token{}, []lexer.Token{})

	toks = []lexer.Token{openParen, foo, closeParen, bar}
	doAssert(toks, []lexer.Token{foo}, []lexer.Token{bar})

	toks = []lexer.Token{openParen, foo, foo, closeParen, bar, bar}
	doAssert(toks, []lexer.Token{foo, foo}, []lexer.Token{bar, bar})

	toks = []lexer.Token{openParen, foo, openParen, bar, closeParen, closeParen}
	doAssert(toks, []lexer.Token{foo, openParen, bar, closeParen}, []lexer.Token{})

	toks = []lexer.Token{openParen, foo, openParen, bar, closeParen, bar, closeParen, foo}
	doAssert(toks, []lexer.Token{foo, openParen, bar, closeParen, bar}, []lexer.Token{foo})
}

func assertParse(t *T, in []lexer.Token, expExpr Expr, expOut []lexer.Token) {
	expr, out, err := parse(in)
	require.Nil(t, err)
	assert.True(t, expExpr.Equal(expr), "expr:%v expExpr:%v", expr, expExpr)
	assert.Equal(t, expOut, out, "out:%v expOut:%v", out, expOut)
}

func TestParseSingle(t *T) {
	foo := lexer.Token{TokenType: lexer.Identifier, Val: "foo"}
	fooExpr := Identifier{tok: tok(foo), ident: "foo"}

	toks := []lexer.Token{foo}
	assertParse(t, toks, fooExpr, []lexer.Token{})

	toks = []lexer.Token{foo, foo}
	assertParse(t, toks, fooExpr, []lexer.Token{foo})

	toks = []lexer.Token{openParen, foo, closeParen, foo}
	assertParse(t, toks, fooExpr, []lexer.Token{foo})

	toks = []lexer.Token{openParen, openParen, foo, closeParen, closeParen, foo}
	assertParse(t, toks, fooExpr, []lexer.Token{foo})
}

func TestParseTuple(t *T) {
	tup := func(ee ...Expr) Expr {
		return Tuple{exprs: ee}
	}

	foo := lexer.Token{TokenType: lexer.Identifier, Val: "foo"}
	fooExpr := Identifier{tok: tok(foo), ident: "foo"}

	toks := []lexer.Token{foo, comma, foo}
	assertParse(t, toks, tup(fooExpr, fooExpr), []lexer.Token{})

	toks = []lexer.Token{foo, comma, foo, foo}
	assertParse(t, toks, tup(fooExpr, fooExpr), []lexer.Token{foo})

	toks = []lexer.Token{foo, comma, foo, comma, foo}
	assertParse(t, toks, tup(fooExpr, fooExpr, fooExpr), []lexer.Token{})

	toks = []lexer.Token{foo, comma, foo, comma, foo, comma, foo}
	assertParse(t, toks, tup(fooExpr, fooExpr, fooExpr, fooExpr), []lexer.Token{})

	toks = []lexer.Token{foo, comma, openParen, foo, comma, foo, closeParen, comma, foo}
	assertParse(t, toks, tup(fooExpr, tup(fooExpr, fooExpr), fooExpr), []lexer.Token{})

	toks = []lexer.Token{foo, comma, openParen, foo, comma, foo, closeParen, comma, foo, foo}
	assertParse(t, toks, tup(fooExpr, tup(fooExpr, fooExpr), fooExpr), []lexer.Token{foo})
}

// This is basically the same as tuple
func TestParsePipe(t *T) {
	mkPipe := func(ee ...Expr) Expr {
		return Pipe{exprs: ee}
	}

	foo := lexer.Token{TokenType: lexer.Identifier, Val: "foo"}
	fooExpr := Identifier{tok: tok(foo), ident: "foo"}

	toks := []lexer.Token{foo, pipe, foo}
	assertParse(t, toks, mkPipe(fooExpr, fooExpr), []lexer.Token{})

	toks = []lexer.Token{foo, pipe, foo, foo}
	assertParse(t, toks, mkPipe(fooExpr, fooExpr), []lexer.Token{foo})

	toks = []lexer.Token{foo, pipe, foo, pipe, foo}
	assertParse(t, toks, mkPipe(fooExpr, fooExpr, fooExpr), []lexer.Token{})

	toks = []lexer.Token{foo, pipe, foo, pipe, foo, pipe, foo}
	assertParse(t, toks, mkPipe(fooExpr, fooExpr, fooExpr, fooExpr), []lexer.Token{})

	toks = []lexer.Token{foo, pipe, openParen, foo, pipe, foo, closeParen, pipe, foo}
	assertParse(t, toks, mkPipe(fooExpr, mkPipe(fooExpr, fooExpr), fooExpr), []lexer.Token{})

	toks = []lexer.Token{foo, pipe, openParen, foo, pipe, foo, closeParen, pipe, foo, foo}
	assertParse(t, toks, mkPipe(fooExpr, mkPipe(fooExpr, fooExpr), fooExpr), []lexer.Token{foo})
}
