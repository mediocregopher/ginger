package gg

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode"
)

// LexerError is returned by Lexer when an unexpected error occurs parsing a
// stream of LexerTokens.
type LexerError struct {
	Err      error
	Row, Col int
}

func (e *LexerError) Error() string {
	return fmt.Sprintf("%d: %d: %s", e.Col, e.Row, e.Err.Error())
}

func (e *LexerError) Unwrap() error {
	return e.Err
}

// LexerTokenKind enumerates the different kinds of LexerToken there can be.
type LexerTokenKind string

// Enumeration of LexerTokenKinds.
const (
	LexerTokenKindName        LexerTokenKind = "name"
	LexerTokenKindNumber      LexerTokenKind = "number"
	LexerTokenKindPunctuation LexerTokenKind = "punctuation"
)

// LexerToken describes a lexigraphical token which is used when deserializing
// Graphs.
type LexerToken struct {
	Kind  LexerTokenKind
	Value string // never empty string

	Row, Col int
}

// Lexer is used to parse a string stream into a sequence of tokens which can
// then be parsed by a Parser.
type Lexer interface {

	// Next will return a LexerToken or a LexerError. io.EOF (wrapped in a
	// LexerError) is returned if the stream being read from is finished.
	Next() (LexerToken, error)
}

type lexer struct {
	r             *bufio.Reader
	stringBuilder *strings.Builder
	err           *LexerError

	// these fields are only needed to keep track of the current "cursor"
	// position when reading.
	lastRow, lastCol int
	prevRune         rune
}

// NewLexer wraps the io.Reader in a Lexer, which will read the io.Reader as a
// sequence of utf-8 characters and parse it into a sequence of LexerTokens.
func NewLexer(r io.Reader) Lexer {
	return &lexer{
		r:             bufio.NewReader(r),
		lastRow:       0,
		lastCol:       -1,
		stringBuilder: new(strings.Builder),
	}
}

// nextRowCol returns the row and column number which the next rune in the
// stream would be at.
func (l *lexer) nextRowCol() (int, int) {

	if l.prevRune == '\n' {
		return l.lastRow + 1, 0
	}

	return l.lastRow, l.lastCol + 1
}

func (l *lexer) fmtErr(err error) *LexerError {

	row, col := l.nextRowCol()

	return &LexerError{
		Err: err,
		Row: row,
		Col: col,
	}
}

func (l *lexer) fmtErrf(str string, args ...interface{}) *LexerError {
	return l.fmtErr(fmt.Errorf(str, args...))
}

// discardRune must _always_ be called only after peekRune.
func (l *lexer) discardRune() {

	r, _, err := l.r.ReadRune()

	if err != nil {
		panic(err)
	}

	l.lastRow, l.lastCol = l.nextRowCol()
	l.prevRune = r
}

func (l *lexer) peekRune() (rune, error) {

	r, _, err := l.r.ReadRune()

	if err != nil {
		return '0', err

	} else if err := l.r.UnreadRune(); err != nil {

		// since the most recent operation on the bufio.Reader was a ReadRune,
		// UnreadRune should never return an error
		panic(err)
	}

	return r, nil
}

// readWhile reads runes until the given predicate returns false, and returns a
// LexerToken of the given kind whose Value is comprised of all runes which
// returned true.
//
// If an error is encountered then both the token (or what's been parsed of it
// so far) and the error are returned.
func (l *lexer) readWhile(
	kind LexerTokenKind, pred func(rune) bool,
) (
	LexerToken, *LexerError,
) {

	row, col := l.nextRowCol()

	l.stringBuilder.Reset()

	var lexErr *LexerError

	for {

		r, err := l.peekRune()

		if err != nil {
			lexErr = l.fmtErrf("peeking next character: %w", err)
			break

		} else if !pred(r) {
			break
		}

		l.stringBuilder.WriteRune(r)

		l.discardRune()
	}

	return LexerToken{
		Kind:  kind,
		Value: l.stringBuilder.String(),
		Row:   row, Col: col,
	}, lexErr
}

// we only support base-10 integers at the moment.
func isNumber(r rune) bool {
	return r == '-' || ('0' <= r && r <= '9')
}

// next can return a token, an error, or both. If an error is returned then no
// further calls to next should occur.
func (l *lexer) next() (LexerToken, *LexerError) {

	for {

		r, err := l.peekRune()

		if err != nil {
			return LexerToken{}, l.fmtErrf("peeking next character: %w", err)
		}

		switch {

		case r == '*': // comment

			// comments are everything up until a newline
			_, err := l.readWhile("", func(r rune) bool {
				return r != '\n'
			})

			if err != nil {
				return LexerToken{}, err
			}

			// terminating newline is deliberately not discarded. Loop and find
			// the next token (which will be that newline).

		case r == '\n':
			// newlines are considered punctuation, not whitespace

			l.discardRune()

			return LexerToken{
				Kind:  LexerTokenKindPunctuation,
				Value: string(r),
				Row:   l.lastRow,
				Col:   l.lastCol,
			}, nil

		case r == '"' || r == '`':

			// reserve double-quote and backtick for string parsing.
			l.discardRune()
			return LexerToken{}, l.fmtErrf("string parsing not yet implemented")

		case unicode.IsLetter(r):
			// letters denote the start of a name

			return l.readWhile(LexerTokenKindName, func(r rune) bool {

				if unicode.In(r, unicode.Letter, unicode.Number, unicode.Mark) {
					return true
				}

				if r == '-' {
					return true
				}

				return false
			})

		case isNumber(r):
			return l.readWhile(LexerTokenKindNumber, isNumber)

		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			// symbols are also considered punctuation

			l.discardRune()

			return LexerToken{
				Kind:  LexerTokenKindPunctuation,
				Value: string(r),
				Row:   l.lastRow,
				Col:   l.lastCol,
			}, nil

		case unicode.IsSpace(r):
			l.discardRune()

		default:
			return LexerToken{}, l.fmtErrf("unexpected character %q", r)
		}

	}
}

func (l *lexer) Next() (LexerToken, error) {

	if l.err != nil {
		return LexerToken{}, l.err
	}

	tok, err := l.next()

	if err != nil {

		l.err = err

		if tok.Kind == "" {
			return LexerToken{}, l.err
		}
	}

	return tok, nil
}
