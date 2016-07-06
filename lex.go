package ginger

import (
	"io"
	"strings"

	"github.com/mediocregopher/lexgo"
)

const (
	number lexgo.TokenType = lexgo.UserDefined + iota
	identifier
	punctuation
)

var numberSet = "0123456789"
var whitespaceSet = " \n\r\t\v\f"
var punctuationSet = ",{}()<>|"

func newLexer(r io.Reader) *lexgo.Lexer {
	return lexgo.NewLexer(r, lexWhitespace)
}

func lexWhitespace(lexer *lexgo.Lexer) lexgo.LexerFunc {
	r, err := lexer.ReadRune()
	if err != nil {
		return nil
	}

	if strings.ContainsRune(whitespaceSet, r) {
		return lexWhitespace
	}

	if r == '/' {
		n, err := lexer.PeekRune()
		if err != nil {
			return nil
		}

		var lexComment func(*lexgo.Lexer) bool
		if n == '/' {
			lexComment = lexLineComment
		} else if n == '*' {
			lexComment = lexBlockComment
		}
		if lexComment != nil {
			if !lexComment(lexer) {
				return nil
			}
			return lexWhitespace
		}
	}

	lexer.BufferRune(r)

	switch {
	case strings.ContainsRune(punctuationSet, r):
		return lexPunctuation
	case strings.ContainsRune(numberSet, r):
		return lexNumber
	default:
		return lexIdentifier
	}
}

// assumes the punctuation has already been buffered
func lexPunctuation(lexer *lexgo.Lexer) lexgo.LexerFunc {
	lexer.Emit(punctuation)
	return lexWhitespace
}

func lexGeneralExpr(lexer *lexgo.Lexer, typ lexgo.TokenType) lexgo.LexerFunc {
	for {
		r, err := lexer.ReadRune()
		if err != nil {
			return nil
		}

		if strings.ContainsRune(whitespaceSet, r) {
			lexer.Emit(typ)
			return lexWhitespace
		}

		if strings.ContainsRune(punctuationSet, r) {
			lexer.Emit(typ)
			lexer.BufferRune(r)
			return lexPunctuation
		}

		lexer.BufferRune(r)
	}
}

func lexNumber(lexer *lexgo.Lexer) lexgo.LexerFunc {
	return lexGeneralExpr(lexer, number)
}

func lexIdentifier(lexer *lexgo.Lexer) lexgo.LexerFunc {
	return lexGeneralExpr(lexer, identifier)
}

func lexLineComment(lexer *lexgo.Lexer) bool {
	for {
		r, err := lexer.ReadRune()
		if err != nil {
			return false
		} else if r == '\n' {
			return true
		}
	}
}

func lexBlockComment(lexer *lexgo.Lexer) bool {
	for {
		r, err := lexer.ReadRune()
		if err != nil {
			return false
		}

		if r == '*' || r == '/' {
			n, err := lexer.PeekRune()
			if err != nil {
				return false
			}
			if r == '*' && n == '/' {
				_, err = lexer.ReadRune()
				return err == nil
			}
			if r == '/' && n == '*' {
				if !lexBlockComment(lexer) {
					return false
				}
			}
		}
	}
}
