package expr

//import . "testing"

//func TestSliceEnclosedToks(t *T) {
//	doAssert := func(in, expOut, expRem []lexer.Token) {
//		out, rem, err := sliceEnclosedToks(in, openParen, closeParen)
//		require.Nil(t, err)
//		assert.Equal(t, expOut, out)
//		assert.Equal(t, expRem, rem)
//	}
//	foo := lexer.Token{TokenType: lexer.Identifier, Val: "foo"}
//	bar := lexer.Token{TokenType: lexer.Identifier, Val: "bar"}
//
//	toks := []lexer.Token{openParen, closeParen}
//	doAssert(toks, []lexer.Token{}, []lexer.Token{})
//
//	toks = []lexer.Token{openParen, foo, closeParen, bar}
//	doAssert(toks, []lexer.Token{foo}, []lexer.Token{bar})
//
//	toks = []lexer.Token{openParen, foo, foo, closeParen, bar, bar}
//	doAssert(toks, []lexer.Token{foo, foo}, []lexer.Token{bar, bar})
//
//	toks = []lexer.Token{openParen, foo, openParen, bar, closeParen, closeParen}
//	doAssert(toks, []lexer.Token{foo, openParen, bar, closeParen}, []lexer.Token{})
//
//	toks = []lexer.Token{openParen, foo, openParen, bar, closeParen, bar, closeParen, foo}
//	doAssert(toks, []lexer.Token{foo, openParen, bar, closeParen, bar}, []lexer.Token{foo})
//}
//
//func assertParse(t *T, in []lexer.Token, expExpr Expr, expOut []lexer.Token) {
//	expr, out, err := parse(in)
//	require.Nil(t, err)
//	assert.True(t, expExpr.equal(expr), "expr:%+v expExpr:%+v", expr, expExpr)
//	assert.Equal(t, expOut, out, "out:%v expOut:%v", out, expOut)
//}
//
//func TestParseSingle(t *T) {
//	foo := lexer.Token{TokenType: lexer.Identifier, Val: "foo"}
//	fooM := lexer.Token{TokenType: lexer.Identifier, Val: "%foo"}
//	fooExpr := Expr{Actual: Identifier("foo")}
//	fooMExpr := Expr{Actual: Macro("foo")}
//
//	toks := []lexer.Token{foo}
//	assertParse(t, toks, fooExpr, []lexer.Token{})
//
//	toks = []lexer.Token{foo, foo}
//	assertParse(t, toks, fooExpr, []lexer.Token{foo})
//
//	toks = []lexer.Token{openParen, foo, closeParen, foo}
//	assertParse(t, toks, fooExpr, []lexer.Token{foo})
//
//	toks = []lexer.Token{openParen, openParen, foo, closeParen, closeParen, foo}
//	assertParse(t, toks, fooExpr, []lexer.Token{foo})
//
//	toks = []lexer.Token{fooM, foo}
//	assertParse(t, toks, fooMExpr, []lexer.Token{foo})
//}
//
//func TestParseTuple(t *T) {
//	tup := func(ee ...Expr) Expr {
//		return Expr{Actual: Tuple(ee)}
//	}
//
//	foo := lexer.Token{TokenType: lexer.Identifier, Val: "foo"}
//	fooExpr := Expr{Actual: Identifier("foo")}
//
//	toks := []lexer.Token{foo, comma, foo}
//	assertParse(t, toks, tup(fooExpr, fooExpr), []lexer.Token{})
//
//	toks = []lexer.Token{foo, comma, foo, foo}
//	assertParse(t, toks, tup(fooExpr, fooExpr), []lexer.Token{foo})
//
//	toks = []lexer.Token{foo, comma, foo, comma, foo}
//	assertParse(t, toks, tup(fooExpr, fooExpr, fooExpr), []lexer.Token{})
//
//	toks = []lexer.Token{foo, comma, foo, comma, foo, comma, foo}
//	assertParse(t, toks, tup(fooExpr, fooExpr, fooExpr, fooExpr), []lexer.Token{})
//
//	toks = []lexer.Token{foo, comma, openParen, foo, comma, foo, closeParen, comma, foo}
//	assertParse(t, toks, tup(fooExpr, tup(fooExpr, fooExpr), fooExpr), []lexer.Token{})
//
//	toks = []lexer.Token{foo, comma, openParen, foo, comma, foo, closeParen, comma, foo, foo}
//	assertParse(t, toks, tup(fooExpr, tup(fooExpr, fooExpr), fooExpr), []lexer.Token{foo})
//}
//
//func TestParseStatement(t *T) {
//	stmt := func(in, to Expr) Expr {
//		return Expr{Actual: Statement{In: in, To: to}}
//	}
//
//	foo := lexer.Token{TokenType: lexer.Identifier, Val: "foo"}
//	fooExpr := Expr{Actual: Identifier("foo")}
//
//	toks := []lexer.Token{foo, arrow, foo}
//	assertParse(t, toks, stmt(fooExpr, fooExpr), []lexer.Token{})
//
//	toks = []lexer.Token{openParen, foo, arrow, foo, closeParen}
//	assertParse(t, toks, stmt(fooExpr, fooExpr), []lexer.Token{})
//
//	toks = []lexer.Token{foo, arrow, openParen, foo, closeParen}
//	assertParse(t, toks, stmt(fooExpr, fooExpr), []lexer.Token{})
//
//	toks = []lexer.Token{foo, arrow, foo}
//	assertParse(t, toks, stmt(fooExpr, fooExpr), []lexer.Token{})
//
//	toks = []lexer.Token{foo, arrow, foo, foo}
//	assertParse(t, toks, stmt(fooExpr, fooExpr), []lexer.Token{foo})
//
//	toks = []lexer.Token{foo, arrow, openParen, foo, closeParen, foo}
//	assertParse(t, toks, stmt(fooExpr, fooExpr), []lexer.Token{foo})
//
//	toks = []lexer.Token{openParen, foo, closeParen, arrow, openParen, foo, closeParen, foo}
//	assertParse(t, toks, stmt(fooExpr, fooExpr), []lexer.Token{foo})
//
//	fooTupExpr := Expr{Actual: Tuple{fooExpr, fooExpr}}
//	toks = []lexer.Token{foo, arrow, openParen, foo, comma, foo, closeParen, foo}
//	assertParse(t, toks, stmt(fooExpr, fooTupExpr), []lexer.Token{foo})
//
//	toks = []lexer.Token{foo, comma, foo, arrow, foo}
//	assertParse(t, toks, stmt(fooTupExpr, fooExpr), []lexer.Token{})
//
//	toks = []lexer.Token{openParen, foo, comma, foo, closeParen, arrow, foo}
//	assertParse(t, toks, stmt(fooTupExpr, fooExpr), []lexer.Token{})
//}
//
//func TestParseBlock(t *T) {
//	stmt := func(in, to Expr) Expr {
//		return Expr{Actual: Statement{In: in, To: to}}
//	}
//	block := func(stmts ...Expr) Expr {
//		return Expr{Actual: Block(stmts)}
//	}
//
//	foo := lexer.Token{TokenType: lexer.Identifier, Val: "foo"}
//	fooExpr := Expr{Actual: Identifier("foo")}
//
//	toks := []lexer.Token{openCurly, foo, arrow, foo, closeCurly}
//	assertParse(t, toks, block(stmt(fooExpr, fooExpr)), []lexer.Token{})
//
//	toks = []lexer.Token{openCurly, foo, arrow, foo, closeCurly, foo}
//	assertParse(t, toks, block(stmt(fooExpr, fooExpr)), []lexer.Token{foo})
//
//	toks = []lexer.Token{openCurly, foo, arrow, foo, openParen, foo, arrow, foo, closeParen, closeCurly, foo}
//	assertParse(t, toks, block(stmt(fooExpr, fooExpr), stmt(fooExpr, fooExpr)), []lexer.Token{foo})
//
//	toks = []lexer.Token{openCurly, foo, arrow, foo, openParen, foo, arrow, foo, closeParen, closeCurly, foo}
//	assertParse(t, toks, block(stmt(fooExpr, fooExpr), stmt(fooExpr, fooExpr)), []lexer.Token{foo})
//}
