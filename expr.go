package ginger

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/mediocregopher/ginger/lexer"
)

// TODO doc strings
// TODO empty blocks
// TODO empty parenthesis

type tok lexer.Token

func (t tok) Token() lexer.Token {
	return lexer.Token(t)
}

type Expr interface {
	Token() lexer.Token
	String() string

	// Equal should return true if the type and value of the other expression
	// are equal. The tokens shouldn't be taken into account
	Equal(Expr) bool
}

////////////////////////////////////////////////////////////////////////////////

type Bool struct {
	tok
	val bool
}

func (b Bool) String() string {
	return fmt.Sprint(b.val)
}

func (b Bool) Equal(e Expr) bool {
	bb, ok := e.(Bool)
	if !ok {
		return false
	}
	return bb.val == b.val
}

////////////////////////////////////////////////////////////////////////////////

type Int struct {
	tok
	val int64
}

func (i Int) String() string {
	return fmt.Sprint(i.val)
}

func (i Int) Equal(e Expr) bool {
	ii, ok := e.(Int)
	if !ok {
		return false
	}
	return ii.val == i.val
}

////////////////////////////////////////////////////////////////////////////////

type String struct {
	tok
	str string
}

func (s String) String() string {
	return strconv.QuoteToASCII(s.str)
}

func (s String) Equal(e Expr) bool {
	ss, ok := e.(String)
	if !ok {
		return false
	}
	return ss.str == s.str
}

////////////////////////////////////////////////////////////////////////////////

type Identifier struct {
	tok
	ident string
}

func (id Identifier) String() string {
	return id.ident
}

func (id Identifier) Equal(e Expr) bool {
	idid, ok := e.(Identifier)
	if !ok {
		return false
	}
	return idid.ident == id.ident
}

////////////////////////////////////////////////////////////////////////////////

type Tuple struct {
	exprs []Expr
}

func (tup Tuple) Token() lexer.Token {
	return tup.exprs[0].Token()
}

func (tup Tuple) String() string {
	strs := make([]string, len(tup.exprs))
	for i := range tup.exprs {
		strs[i] = tup.exprs[i].String()
	}
	return "(" + strings.Join(strs, ", ") + ")"
}

func (tup Tuple) Equal(e Expr) bool {
	tuptup, ok := e.(Tuple)
	if !ok || len(tuptup.exprs) != len(tup.exprs) {
		return false
	}
	for i := range tup.exprs {
		if !tup.exprs[i].Equal(tuptup.exprs[i]) {
			return false
		}
	}
	return true
}

////////////////////////////////////////////////////////////////////////////////

type Pipe struct {
	exprs []Expr
}

func (p Pipe) Token() lexer.Token {
	return p.exprs[0].Token()
}

func (p Pipe) String() string {
	strs := make([]string, len(p.exprs))
	for i := range p.exprs {
		strs[i] = p.exprs[i].String()
	}
	return "(" + strings.Join(strs, "|") + ")"
}

func (p Pipe) Equal(e Expr) bool {
	pp, ok := e.(Pipe)
	if !ok || len(pp.exprs) != len(p.exprs) {
		return false
	}
	for i := range p.exprs {
		if !p.exprs[i].Equal(pp.exprs[i]) {
			return false
		}
	}
	return true
}

////////////////////////////////////////////////////////////////////////////////

type Statement struct {
	in   Expr
	pipe Pipe
}

func (s Statement) Token() lexer.Token {
	return s.in.Token()
}

func (s Statement) String() string {
	return fmt.Sprintf("(%s > %s)", s.in.String(), s.pipe.String())
}

func (s Statement) Equal(e Expr) bool {
	ss, ok := e.(Statement)
	return ok && s.in.Equal(ss.in) && s.pipe.Equal(ss.pipe)
}

////////////////////////////////////////////////////////////////////////////////

type Block struct {
	stmts []Statement
}

func (b Block) Token() lexer.Token {
	return b.stmts[0].Token()
}

func (b Block) String() string {
	strs := make([]string, len(b.stmts))
	for i := range b.stmts {
		strs[i] = b.stmts[i].String()
	}
	return fmt.Sprintf("{ %s }", strings.Join(strs, " "))
}

func (b Block) Equal(e Expr) bool {
	bb, ok := e.(Block)
	if !ok {
		return false
	}
	for i := range b.stmts {
		if !b.stmts[i].Equal(bb.stmts[i]) {
			return false
		}
	}
	return true
}

////////////////////////////////////////////////////////////////////////////////

type exprErr struct {
	reason string
	tok    lexer.Token
	tokCtx string // e.g. "block starting at" or "open paren at"
}

func (e exprErr) Error() string {
	msg := e.reason
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
	openParen  = lexer.Token{TokenType: lexer.Punctuation, Val: "("}
	closeParen = lexer.Token{TokenType: lexer.Punctuation, Val: ")"}
	openCurly  = lexer.Token{TokenType: lexer.Punctuation, Val: "{"}
	closeCurly = lexer.Token{TokenType: lexer.Punctuation, Val: "}"}
	comma      = lexer.Token{TokenType: lexer.Punctuation, Val: ","}
	pipe       = lexer.Token{TokenType: lexer.Punctuation, Val: "|"}
	arrow      = lexer.Token{TokenType: lexer.Punctuation, Val: ">"}
)

func parse(toks []lexer.Token) (Expr, []lexer.Token, error) {
	expr, toks, err := parseSingle(toks)
	if err != nil {
		return nil, nil, err
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
		return nil, nil, exprErr{
			reason: "could not parse token",
			tok:    toks[0],
		}
	}

	if toks[0].Equal(openParen) {
		starter := toks[0]
		var ptoks []lexer.Token
		ptoks, toks, err = sliceEnclosedToks(toks, openParen, closeParen)
		if err != nil {
			return nil, nil, err
		}

		if expr, ptoks, err = parse(ptoks); err != nil {
			return nil, nil, err
		} else if len(ptoks) > 0 {
			return nil, nil, exprErr{
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
			return nil, nil, err
		}

		if expr, err = parseBlock(btoks); err != nil {
			return nil, nil, err
		}
		return expr, toks, nil
	}

	if expr, err = parseNonPunct(toks[0]); err != nil {
		return nil, nil, err
	}
	return expr, toks[1:], nil
}

func parseNonPunct(tok lexer.Token) (Expr, error) {
	if tok.TokenType == lexer.Identifier {
		return parseIdentifier(tok)
	} else if tok.TokenType == lexer.String {
		return parseString(tok)
	}

	return nil, exprErr{
		reason: "unexpected non-punctuation token",
		tok:    tok,
	}
}

func parseIdentifier(t lexer.Token) (Expr, error) {
	if t.Val[0] == '-' || (t.Val[0] >= '0' && t.Val[0] <= '9') {
		n, err := strconv.ParseInt(t.Val, 10, 64)
		if err != nil {
			return nil, exprErr{
				reason: "error parsing number",
				// TODO err: err,
				tok: t,
			}
		}
		return Int{tok: tok(t), val: n}, nil
	}

	if t.Val == "true" {
		return Bool{tok: tok(t), val: true}, nil
	} else if t.Val == "false" {
		return Bool{tok: tok(t), val: false}, nil
	}

	return Identifier{tok: tok(t), ident: t.Val}, nil
}

func parseString(t lexer.Token) (Expr, error) {
	str, err := strconv.Unquote(t.Val)
	if err != nil {
		return nil, exprErr{
			reason: "error parsing string",
			// TODO err: err,
			tok: t,
		}
	}
	return String{tok: tok(t), str: str}, nil
}

func parseConnectingPunct(toks []lexer.Token, root Expr) (Expr, []lexer.Token, error) {
	if toks[0].Equal(comma) {
		return parseTuple(toks, root)

	} else if toks[0].Equal(pipe) {
		return parsePipe(toks, root)

	} else if toks[0].Equal(arrow) {
		expr, toks, err := parse(toks[1:])
		if err != nil {
			return nil, nil, err
		}
		pipe, ok := expr.(Pipe)
		if !ok {
			pipe = Pipe{exprs: []Expr{expr}}
		}
		return Statement{in: root, pipe: pipe}, toks, nil
	}

	return root, toks, nil
}

func parseTuple(toks []lexer.Token, root Expr) (Expr, []lexer.Token, error) {
	rootTup, ok := root.(Tuple)
	if !ok {
		rootTup = Tuple{exprs: []Expr{root}}
	}

	if len(toks) < 2 {
		return rootTup, toks, nil
	} else if !toks[0].Equal(comma) {
		return rootTup, toks, nil
	}

	var expr Expr
	var err error
	if expr, toks, err = parseSingle(toks[1:]); err != nil {
		return nil, nil, err
	}

	rootTup.exprs = append(rootTup.exprs, expr)
	return parseTuple(toks, rootTup)
}

func parsePipe(toks []lexer.Token, root Expr) (Expr, []lexer.Token, error) {
	rootTup, ok := root.(Pipe)
	if !ok {
		rootTup = Pipe{exprs: []Expr{root}}
	}

	if len(toks) < 2 {
		return rootTup, toks, nil
	} else if !toks[0].Equal(pipe) {
		return rootTup, toks, nil
	}

	var expr Expr
	var err error
	if expr, toks, err = parseSingle(toks[1:]); err != nil {
		return nil, nil, err
	}

	rootTup.exprs = append(rootTup.exprs, expr)
	return parsePipe(toks, rootTup)
}

// parseBlock assumes that the given token list is the entire block, already
// pulled from outer curly braces by sliceEnclosedToks, or determined to be the
// entire block in some other way.
func parseBlock(toks []lexer.Token) (Expr, error) {
	b := Block{}

	var expr Expr
	var err error
	for {
		if len(toks) == 0 {
			return b, nil
		}

		if expr, toks, err = parse(toks); err != nil {
			return nil, err
		}
		stmt, ok := expr.(Statement)
		if !ok {
			return nil, exprErr{
				reason: "blocks may only contain full statements",
				tok:    expr.Token(),
				tokCtx: "non-statement here",
			}
		}
		b.stmts = append(b.stmts, stmt)
	}
}
