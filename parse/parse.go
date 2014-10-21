// The parse package implements a syntax parser for the ginger syntax. It can
// read in any io.Reader and returns fully parsed Elem's from the types package
// that it finds.
package parse

import (
	"bytes"
	"io"
	"fmt"
	"strconv"
	"unsafe"

	"github.com/mediocregopher/ginger/parse/lex"
	"github.com/mediocregopher/ginger/seq"
	"github.com/mediocregopher/ginger/types"
)

const int_bits = int(unsafe.Sizeof(int(0)) * 8)

var closers = map[string]string{
	"(": ")",
	"[": "]",
	"{": "}",
}

// The lexer only indicates a bare string, but technically an integer or a float
// is a bare string so we must try and convert to one of those first
func parseBareString(tok *lex.Token) types.Elem {
	if i, err := strconv.ParseInt(tok.Val, 10, int_bits); err == nil {
		return types.GoType{int(i)}

	} else if int_bits == 64 {
		// We don't want to bother with the next case if int_bits is 64

	} else if i64, err := strconv.ParseInt(tok.Val, 10, 64); err == nil {
		return types.GoType{int64(i64)}
	}

	if f32, err := strconv.ParseFloat(tok.Val, 32); err == nil {
		return types.GoType{float32(f32)}
	}

	if f64, err := strconv.ParseFloat(tok.Val, 64); err == nil {
		return types.GoType{float64(f64)}
	}

	if tok.Val[0] != ':' {
		return types.GoType{":"+tok.Val}
	}

	return types.GoType{tok.Val}
}

func parseQuotedString(tok *lex.Token) (types.Elem, error) {
	s, err := strconv.Unquote(tok.Val)
	if err != nil {
		return nil, err
	}

	return types.GoType{s}, nil
}

type Parser struct {
	l *lex.Lexer
}

// Returns a NewParser, using the lex package as the tokenizer
func NewParser(r io.Reader) *Parser {
	p := Parser{
		l: lex.NewLexer(r),
	}
	return &p
}

// Reads a full element, and any sub-elements (if the top-level element is a
// data-structure) into an Elem and returns it. Returns any errors, including
// io.EOF, if it runs into them instead
func (p *Parser) ReadElem() (types.Elem, error) {
	tok := p.l.Next()
	return p.parseToken(tok)
}

func (p *Parser) parseToken(tok *lex.Token) (types.Elem, error) {
	if tok == nil {
		return nil, io.EOF
	}

	switch tok.Type {
	case lex.Err:
		return nil, tok.AsError()
	case lex.BareString:
		return parseBareString(tok), nil
	case lex.QuotedString:
		return parseQuotedString(tok)
	case lex.Open:
		series, err := p.readUntil(closers[tok.Val])
		if err != nil {
			return nil, err
		}
		if tok.Val == "(" {
			return seq.NewList(series...), nil
		} else if tok.Val == "{" {
			if len(series) % 2 != 0 {
				return nil, fmt.Errorf("hash must have even number of elements")
			}
			kvs := make([]*seq.KV, 0, len(series) / 2)
			for i := 0; i < len(series); i += 2 {
				kv := seq.KV{series[i], series[i+1]}
				kvs = append(kvs, &kv)
			}
			return seq.NewHashMap(kvs...), nil
		}

		panic("should never get here")
	
	default:
		return nil, fmt.Errorf("Unexpected %q", tok.Val)
	}
}


func (p *Parser) readUntil(closer string) ([]types.Elem, error) {
	series := make([]types.Elem, 0, 4)
	for {
		tok := p.l.Next()
		switch err := tok.AsError(); err {
		case nil:
		case io.EOF:
			return nil, fmt.Errorf("Unexpected EOF")
		default:
			return nil, err
		}

		if tok.Type != lex.Close {
			e, err := p.parseToken(tok)
			if err != nil {
				return nil, err
			}
			series = append(series, e)
			continue
		}

		if tok.Val != closer {
			return nil, fmt.Errorf("Unexpected %q", tok.Val)
		}

		return series, nil
	}
}

// Parses the first Elem it finds out of the given string and returns it
func ParseString(input string) (types.Elem, error) {
	buf := bytes.NewBufferString(input)
	p := NewParser(buf)
	return p.ReadElem()
}
