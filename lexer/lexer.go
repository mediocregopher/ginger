package lexer

import (
	"bufio"
	"bytes"
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
)

// Token is a single token which has been read in. All Tokens have a non-empty
// Val
type Token struct {
	TokenType
	Val      string
	Row, Col int
}

type lexerFn func(*Lexer, rune, rune) lexerFn

// Lexer is used to read in ginger tokens from a source. HasNext() must be
// called before every call to Next(), and Err() must be called once HasNext()
// returns false.
type Lexer struct {
	in  *bufio.Reader
	out *bytes.Buffer
	cur lexerFn

	next []Token
	err  error

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

	l.next = append(l.next, Token{
		TokenType: t,
		Val:       str,
		Row:       l.row,
		Col:       l.col,
	})
	l.row = -1
	l.col = -1
}

func (l *Lexer) readRune() (rune, bool) {
	r, _, err := l.in.ReadRune()
	if err != nil {
		l.err = err
		return r, false
	}
	return r, true
}

func (l *Lexer) peekRune() (rune, bool) {
	r, ok := l.readRune()
	if !ok {
		return r, ok
	}

	if err := l.in.UnreadRune(); err != nil {
		l.err = err
		return r, false
	}
	return r, true
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
	if l.err != nil || l.cur == nil {
		return false
	}

	for {
		if len(l.next) > 0 {
			return true
		}

		var ok bool
		var r, n rune
		if r, ok = l.readRune(); !ok {
			return false
		}

		if n, ok = l.peekRune(); !ok {
			return false
		}

		if r == '\n' {
			l.absRow++
			l.absCol = 0
		} else {
			l.absCol++
		}

		l.cur = l.cur(l, r, n)
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

// Err returns the error which caused HasNext to return false. Will return nil
// if the error was io.EOF
func (l *Lexer) Err() error {
	if l.err != nil && l.err != io.EOF {
		return l.err
	} else if l.out.Len() > 0 {
		return fmt.Errorf("incomplete token: %q", l.out.String())
	}
	return nil
}

var whitespaceSet = " \n\r\t\v\f"
var punctuationSet = ",{}()<>|"
var identifierSepSet = whitespaceSet + punctuationSet

func lex(lexer *Lexer, r, n rune) lexerFn {
	switch {
	case strings.ContainsRune(whitespaceSet, r):
		return lex
	case r == '/' && n == '/':
		return lexLineComment
	case strings.ContainsRune(punctuationSet, r):
		return lexPunctuation(lexer, r, n)
	case r == '"' || r == '\'' || r == '`':
		canEscape := r != '`'
		return lexStrStart(lexer, r, makeLexStr(r, canEscape))
	default:
		return lexIdentifier(lexer, r, n)
	}
}

func lexPunctuation(lexer *Lexer, r, n rune) lexerFn {
	lexer.bufferRune(r)
	lexer.emit(Punctuation)
	return lex
}

func lexIdentifier(lexer *Lexer, r, n rune) lexerFn {
	if strings.ContainsRune(identifierSepSet, r) {
		lexer.emit(Identifier)
		return lex(lexer, r, n)
	}

	lexer.bufferRune(r)
	return lexIdentifier
}

func lexLineComment(lexer *Lexer, r, n rune) lexerFn {
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
	fn = func(lexer *Lexer, r, n rune) lexerFn {
		if canEscape && r == '\\' && n == quoteC {
			lexer.bufferRune(r)
			lexer.bufferRune(n)
			return lexSkipThen(fn)
		}

		lexer.bufferRune(r)
		if r == quoteC {
			lexer.emit(String)
			return lex
		}

		return fn
	}
	return fn
}

func lexSkipThen(then lexerFn) lexerFn {
	return func(lexer *Lexer, r, n rune) lexerFn {
		return then
	}
}
