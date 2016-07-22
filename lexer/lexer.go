package lexer

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
)

// TokenType indicates the type of a token
type TokenType string

// Different token types
const (
	Identifier  TokenType = "identifier"
	Punctuation TokenType = "punctuation"
	String      TokenType = "string"
	Err         TokenType = "err"
	EOF         TokenType = "eof"
)

// Token is a single token which has been read in. All Tokens have a non-empty
// Val
type Token struct {
	TokenType
	Val      string
	Row, Col int
}

// Equal returns whether two tokens are of equal type and value
func (tok Token) Equal(tok2 Token) bool {
	return tok.TokenType == tok2.TokenType && tok.Val == tok2.Val
}

// Err returns the error contained by the token, if any. Only returns non-nil if
// TokenType is Err or EOR
func (tok Token) Err() error {
	if tok.TokenType == Err || tok.TokenType == EOF {
		return errors.New(tok.Val)
	}
	return nil
}

func (tok Token) String() string {
	var typ string
	switch tok.TokenType {
	case Identifier:
		typ = "ident"
	case Punctuation:
		typ = "punct"
	case String:
		typ = "str"
	case Err, EOF:
		typ = "err"
	}
	return fmt.Sprintf("%s(%q)", typ, tok.Val)
}

type lexerFn func(*Lexer) lexerFn

// Lexer is used to read in ginger tokens from a source. HasNext() must be
// called before every call to Next()
type Lexer struct {
	in  *bufio.Reader
	out *bytes.Buffer
	cur lexerFn

	next []Token

	row, col       int
	absRow, absCol int
}

// New returns a Lexer which will read tokens from the given source.
func New(r io.Reader) *Lexer {
	return &Lexer{
		in:  bufio.NewReader(r),
		out: new(bytes.Buffer),
		cur: lex,

		row: -1,
		col: -1,
	}
}

func (l *Lexer) emit(t TokenType) {
	str := l.out.String()
	if str == "" {
		panic("cannot emit empty token")
	}
	l.out.Reset()

	l.emitTok(Token{
		TokenType: t,
		Val:       str,
		Row:       l.row,
		Col:       l.col,
	})
}

func (l *Lexer) emitErr(err error) {
	tok := Token{
		TokenType: Err,
		Val:       err.Error(),
		Row:       l.absRow,
		Col:       l.absCol,
	}
	if err == io.EOF {
		tok.TokenType = EOF
	}
	l.emitTok(tok)
}

func (l *Lexer) emitTok(tok Token) {
	l.next = append(l.next, tok)
	l.row = -1
	l.col = -1
}

func (l *Lexer) readRune() (rune, error) {
	r, _, err := l.in.ReadRune()
	if err != nil {
		return r, err
	}

	if r == '\n' {
		l.absRow++
		l.absCol = 0
	} else {
		l.absCol++
	}

	return r, err
}

func (l *Lexer) peekRune() (rune, error) {
	r, _, err := l.in.ReadRune()
	if err != nil {
		return r, err
	}

	if err := l.in.UnreadRune(); err != nil {
		return r, err
	}
	return r, nil
}

func (l *Lexer) readAndPeek() (rune, rune, error) {
	r, err := l.readRune()
	if err != nil {
		return r, 0, err
	}

	n, err := l.peekRune()
	return r, n, err
}

func (l *Lexer) bufferRune(r rune) {
	l.out.WriteRune(r)
	if l.row < 0 && l.col < 0 {
		l.row, l.col = l.absRow, l.absCol
	}
}

// HasNext returns true if Next should be called, and false if it should not be
// called and Err should be called instead. When HasNext returns false the Lexer
// is considered to be done
func (l *Lexer) HasNext() bool {
	for {
		if len(l.next) > 0 {
			return true
		} else if l.cur == nil {
			return false
		}
		l.cur = l.cur(l)
	}
}

// Next returns the next available token. HasNext must be called before every
// call to Next
func (l *Lexer) Next() Token {
	t := l.next[0]
	l.next = l.next[1:]
	if len(l.next) == 0 {
		l.next = nil
	}
	return t
}

var whitespaceSet = " \n\r\t\v\f"
var punctuationSet = ",{}()<>|"
var identifierSepSet = whitespaceSet + punctuationSet

func lex(l *Lexer) lexerFn {
	r, err := l.readRune()
	if err != nil {
		l.emitErr(err)
		return nil
	}

	// handle comments first, cause we have to peek for those. We ignore errors,
	// and assume that any error that would happen here will happen again the
	// next read
	if n, _ := l.peekRune(); n == '/' {
		return lexLineComment
	}

	return lexSingleRune(l, r)
}

func lexSingleRune(l *Lexer, r rune) lexerFn {
	switch {
	case strings.ContainsRune(whitespaceSet, r):
		return lex
	case strings.ContainsRune(punctuationSet, r):
		l.bufferRune(r)
		l.emit(Punctuation)
		return lex
	case r == '"' || r == '\'' || r == '`':
		canEscape := r != '`'
		return lexStrStart(l, r, makeLexStr(r, canEscape))
	default:
		l.bufferRune(r)
		return lexIdentifier
	}
}

func lexIdentifier(l *Lexer) lexerFn {
	r, err := l.readRune()
	if err != nil {
		l.emit(Identifier)
		l.emitErr(err)
		return nil
	}

	if strings.ContainsRune(identifierSepSet, r) {
		l.emit(Identifier)
		return lexSingleRune(l, r)
	}

	l.bufferRune(r)

	return lexIdentifier
}

func lexLineComment(l *Lexer) lexerFn {
	r, err := l.readRune()
	if err != nil {
		l.emitErr(err)
		return nil
	}
	if r == '\n' {
		return lex
	}
	return lexLineComment
}

func lexStrStart(lexer *Lexer, r rune, then lexerFn) lexerFn {
	lexer.bufferRune(r)
	return then
}

func makeLexStr(quoteC rune, canEscape bool) lexerFn {
	var fn lexerFn
	fn = func(l *Lexer) lexerFn {
		r, n, err := l.readAndPeek()
		if err != nil {
			if err == io.EOF {
				if r == quoteC {
					l.bufferRune(r)
					l.emit(String)
					l.emitErr(err)
					return nil
				}
				l.emitErr(errors.New("expected end of string, got end of file"))
				return nil
			}
		}

		if canEscape && r == '\\' && n == quoteC {
			l.bufferRune(r)
			l.bufferRune(n)
			return lexSkipThen(fn)
		}

		l.bufferRune(r)
		if r == quoteC {
			l.emit(String)
			return lex
		}

		return fn
	}
	return fn
}

func lexSkipThen(then lexerFn) lexerFn {
	return func(l *Lexer) lexerFn {
		if _, err := l.readRune(); err != nil {
			l.emitErr(err)
			return nil
		}
		return then
	}
}
