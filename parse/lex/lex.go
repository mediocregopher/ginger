// The lex package implements a lexical reader which can take in any io.Reader.
// It does not care about the meaning or logical validity of the tokens it
// parses out, it simply does its job.
package lex

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"unicode"
)

type TokenType int

const (
	BareString TokenType = iota
	QuotedString
	Open
	Close
	Err
	eof
)

var invalidBareStringRunes = map[rune]bool{
	'"':  true,
	'\'': true,
	'(':  true,
	')':  true,
	'[':  true,
	']':  true,
	'{':  true,
	'}':  true,
}

// Token represents a single set of characters which *could* be a valid token of
// the given type
type Token struct {
	Type TokenType
	Val  string
}

// Returns the token's value as an error, or nil if the token is not of type
// Err. If the token is nil returns io.EOF, since that is the ostensible meaning
func (t *Token) AsError() error {
	if t == nil {
		return io.EOF
	}
	if t.Type != Err {
		return nil
	}
	return errors.New(t.Val)
}

var (
	errInvalidUTF8 = errors.New("invalid utf8 character")
)

// Lexer reads through an io.Reader and emits Tokens from it.
type Lexer struct {
	r      *bufio.Reader
	outbuf *bytes.Buffer
	ch     chan *Token
}

// NewLexer constructs a new Lexer struct and returns it. r is internally
// wrapped with a bufio.Reader, unless it already is one. This will spawn a
// go-routine which reads from r until it hits an error, at which point it will
// end execution.
func NewLexer(r io.Reader) *Lexer {
	var br *bufio.Reader
	var ok bool
	if br, ok = r.(*bufio.Reader); !ok {
		br = bufio.NewReader(r)
	}

	l := Lexer{
		r:      br,
		ch:     make(chan *Token),
		outbuf: bytes.NewBuffer(make([]byte, 0, 1024)),
	}

	go l.spin()

	return &l
}

func (l *Lexer) spin() {
	f := lexWhitespace
	for {
		f = f(l)
		if f == nil {
			return
		}
	}
}

// Returns the next available token, or nil if EOF has been reached. If an error
// other than EOF has been reached it will be returned as the Err token type,
// and this method should not be called again after that.
func (l *Lexer) Next() *Token {
	t := <-l.ch
	if t.Type == eof {
		return nil
	}
	return t
}

func (l *Lexer) emit(t TokenType) {
	str := l.outbuf.String()
	l.ch <- &Token{
		Type: t,
		Val:  str,
	}
	l.outbuf.Reset()
}

func (l *Lexer) peek() (rune, error) {
	r, err := l.readRune()
	if err != nil {
		return 0, err
	}
	if err = l.r.UnreadRune(); err != nil {
		return 0, err
	}
	return r, nil
}

func (l *Lexer) readRune() (rune, error) {
	r, i, err := l.r.ReadRune()
	if err != nil {
		return 0, err
	} else if r == unicode.ReplacementChar && i == 1 {
		return 0, errInvalidUTF8
	}
	return r, nil
}

func (l *Lexer) err(err error) lexerFunc {
	if err == io.EOF {
		l.ch <- &Token{eof, ""}
	} else {
		l.ch <- &Token{Err, err.Error()}
	}
	close(l.ch)
	return nil
}

func (l *Lexer) errf(format string, args ...interface{}) lexerFunc {
	s := fmt.Sprintf(format, args...)
	l.ch <- &Token{Err, s}
	close(l.ch)
	return nil
}

type lexerFunc func(*Lexer) lexerFunc

func lexWhitespace(l *Lexer) lexerFunc {
	r, err := l.readRune()
	if err != nil {
		return l.err(err)
	}

	if unicode.IsSpace(r) {
		return lexWhitespace
	}

	l.outbuf.WriteRune(r)

	switch r {
	case '"':
		return lexQuotedString
	case '(':
		l.emit(Open)
	case ')':
		l.emit(Close)
	case '[':
		l.emit(Open)
	case ']':
		l.emit(Close)
	case '{':
		l.emit(Open)
	case '}':
		l.emit(Close)
	default:
		return lexBareString
	}

	return lexWhitespace
}

func lexQuotedString(l *Lexer) lexerFunc {
	r, err := l.readRune()
	if err != nil {
		l.emit(QuotedString)
		return l.err(err)
	}

	l.outbuf.WriteRune(r)
	buf := l.outbuf.Bytes()

	if r == '"' && buf[len(buf)-2] != '\\' {
		l.emit(QuotedString)
		return lexWhitespace
	}
	return lexQuotedString
}

func lexBareString(l *Lexer) lexerFunc {
	r, err := l.peek()
	if err != nil {
		l.emit(BareString)
		return l.err(err)
	}

	if _, ok := invalidBareStringRunes[r]; ok || unicode.IsSpace(r) {
		l.emit(BareString)
		return lexWhitespace
	}

	if _, err = l.readRune(); err != nil {
		l.emit(BareString)
		return l.err(err)
	}

	l.outbuf.WriteRune(r)
	return lexBareString
}
