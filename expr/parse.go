package expr

import (
	"fmt"
	"io"
	"strconv"

	"github.com/mediocregopher/ginger/lexer"
)

type exprErr struct {
	reason string
	err    error
	tok    lexer.Token
	tokCtx string // e.g. "block starting at" or "open paren at"
}

func (e exprErr) Error() string {
	var msg string
	if e.err != nil {
		msg = e.err.Error()
	} else {
		msg = e.reason
	}
	if err := e.tok.Err(); err != nil {
		msg += " - token error: " + err.Error()
	} else if (e.tok != lexer.Token{}) {
		msg += " - "
		if e.tokCtx != "" {
			msg += e.tokCtx + ": "
		}
		msg = fmt.Sprintf("%s [line:%d col:%d]", msg, e.tok.Row, e.tok.Col)
	}
	return msg
}

////////////////////////////////////////////////////////////////////////////////

// toks[0] must be start
func sliceEnclosedToks(toks []lexer.Token, start, end lexer.Token) ([]lexer.Token, []lexer.Token, error) {
	c := 1
	ret := []lexer.Token{}
	first := toks[0]
	for i, tok := range toks[1:] {
		if tok.Err() != nil {
			return nil, nil, exprErr{
				reason: fmt.Sprintf("missing closing %v", end),
				tok:    tok,
			}
		}

		if tok.Equal(start) {
			c++
		} else if tok.Equal(end) {
			c--
		}
		if c == 0 {
			return ret, toks[2+i:], nil
		}
		ret = append(ret, tok)
	}

	return nil, nil, exprErr{
		reason: fmt.Sprintf("missing closing %v", end),
		tok:    first,
		tokCtx: "starting at",
	}
}

// Parse reads in all expressions it can from the given io.Reader and returns
// them
func Parse(r io.Reader) ([]Expr, error) {
	toks := readAllToks(r)
	var ret []Expr
	var expr Expr
	var err error
	for len(toks) > 0 {
		if toks[0].TokenType == lexer.EOF {
			return ret, nil
		}
		expr, toks, err = parse(toks)
		if err != nil {
			return nil, err
		}
		ret = append(ret, expr)
	}
	return ret, nil
}

// ParseAsBlock reads the given io.Reader as if it was implicitly surrounded by
// curly braces, making it into a Block. This means all expressions from the
// io.Reader *must* be statements. The returned Expr's Actual will always be a
// Block.
func ParseAsBlock(r io.Reader) (Expr, error) {
	return parseBlock(readAllToks(r))
}

func readAllToks(r io.Reader) []lexer.Token {
	l := lexer.New(r)
	var toks []lexer.Token
	for l.HasNext() {
		toks = append(toks, l.Next())
	}
	return toks
}

// For all parse methods it is assumed that toks is not empty

var (
	openParen  = lexer.Token{TokenType: lexer.Wrapper, Val: "("}
	closeParen = lexer.Token{TokenType: lexer.Wrapper, Val: ")"}
	openCurly  = lexer.Token{TokenType: lexer.Wrapper, Val: "{"}
	closeCurly = lexer.Token{TokenType: lexer.Wrapper, Val: "}"}
	comma      = lexer.Token{TokenType: lexer.Punctuation, Val: ","}
	arrow      = lexer.Token{TokenType: lexer.Punctuation, Val: ">"}
)

func parse(toks []lexer.Token) (Expr, []lexer.Token, error) {
	expr, toks, err := parseSingle(toks)
	if err != nil {
		return Expr{}, nil, err
	}

	if len(toks) > 0 && toks[0].TokenType == lexer.Punctuation {
		return parseConnectingPunct(toks, expr)
	}

	return expr, toks, nil
}

func parseSingle(toks []lexer.Token) (Expr, []lexer.Token, error) {
	var expr Expr
	var err error

	if toks[0].Err() != nil {
		return Expr{}, nil, exprErr{
			reason: "could not parse token",
			tok:    toks[0],
		}
	}

	if toks[0].Equal(openParen) {
		starter := toks[0]
		var ptoks []lexer.Token
		ptoks, toks, err = sliceEnclosedToks(toks, openParen, closeParen)
		if err != nil {
			return Expr{}, nil, err
		}

		if expr, ptoks, err = parse(ptoks); err != nil {
			return Expr{}, nil, err
		} else if len(ptoks) > 0 {
			return Expr{}, nil, exprErr{
				reason: "multiple expressions inside parenthesis",
				tok:    starter,
				tokCtx: "starting at",
			}
		}
		return expr, toks, nil

	} else if toks[0].Equal(openCurly) {
		var btoks []lexer.Token
		btoks, toks, err = sliceEnclosedToks(toks, openCurly, closeCurly)
		if err != nil {
			return Expr{}, nil, err
		}

		if expr, err = parseBlock(btoks); err != nil {
			return Expr{}, nil, err
		}
		return expr, toks, nil
	}

	if expr, err = parseNonPunct(toks[0]); err != nil {
		return Expr{}, nil, err
	}
	return expr, toks[1:], nil
}

func parseNonPunct(tok lexer.Token) (Expr, error) {
	if tok.TokenType == lexer.Identifier {
		return parseIdentifier(tok)
	} else if tok.TokenType == lexer.String {
		return parseString(tok)
	}

	return Expr{}, exprErr{
		reason: "unexpected non-punctuation token",
		tok:    tok,
	}
}

func parseIdentifier(t lexer.Token) (Expr, error) {
	e := Expr{Token: t}
	if t.Val[0] == '-' || (t.Val[0] >= '0' && t.Val[0] <= '9') {
		n, err := strconv.ParseInt(t.Val, 10, 64)
		if err != nil {
			return Expr{}, exprErr{
				err: err,
				tok: t,
			}
		}
		e.Actual = Int(n)

	} else if t.Val == "%true" {
		e.Actual = Bool(true)

	} else if t.Val == "%false" {
		e.Actual = Bool(false)

	} else if t.Val[0] == '%' {
		e.Actual = Macro(t.Val[1:])

	} else {
		e.Actual = Identifier(t.Val)
	}

	return e, nil
}

func parseString(t lexer.Token) (Expr, error) {
	str, err := strconv.Unquote(t.Val)
	if err != nil {
		return Expr{}, exprErr{
			err: err,
			tok: t,
		}
	}
	return Expr{Token: t, Actual: String(str)}, nil
}

func parseConnectingPunct(toks []lexer.Token, root Expr) (Expr, []lexer.Token, error) {
	if toks[0].Equal(comma) {
		return parseTuple(toks, root)

	} else if toks[0].Equal(arrow) {
		expr, toks, err := parse(toks[1:])
		if err != nil {
			return Expr{}, nil, err
		}
		return Expr{Token: root.Token, Actual: Statement{In: root, To: expr}}, toks, nil
	}

	return root, toks, nil
}

func parseTuple(toks []lexer.Token, root Expr) (Expr, []lexer.Token, error) {
	rootTup, ok := root.Actual.(Tuple)
	if !ok {
		rootTup = Tuple{root}
	}

	// rootTup is modified throughout, be we need to make it into an Expr for
	// every return, which is annoying. so make a function to do it on the fly
	mkRoot := func() Expr {
		return Expr{Token: rootTup[0].Token, Actual: rootTup}
	}

	if len(toks) < 2 {
		return mkRoot(), toks, nil
	} else if !toks[0].Equal(comma) {
		if toks[0].TokenType == lexer.Punctuation {
			return parseConnectingPunct(toks, mkRoot())
		}
		return mkRoot(), toks, nil
	}

	var expr Expr
	var err error
	if expr, toks, err = parseSingle(toks[1:]); err != nil {
		return Expr{}, nil, err
	}

	rootTup = append(rootTup, expr)
	return parseTuple(toks, mkRoot())
}

// parseBlock assumes that the given token list is the entire block, already
// pulled from outer curly braces by sliceEnclosedToks, or determined to be the
// entire block in some other way.
func parseBlock(toks []lexer.Token) (Expr, error) {
	b := Block{}
	first := toks[0]
	var expr Expr
	var err error
	for {
		if len(toks) == 0 {
			return Expr{Token: first, Actual: b}, nil
		}

		if expr, toks, err = parse(toks); err != nil {
			return Expr{}, err
		}
		if _, ok := expr.Actual.(Statement); !ok {
			return Expr{}, exprErr{
				reason: "blocks may only contain full statements",
				tok:    expr.Token,
				tokCtx: "non-statement here",
			}
		}
		b = append(b, expr)
	}
}
