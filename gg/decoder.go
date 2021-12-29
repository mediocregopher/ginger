package gg

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/mediocregopher/ginger/graph"
)

// Punctuations which are used in the gg file format.
const (
	punctTerm       = ";"
	punctOp         = "<"
	punctAssign     = "="
	punctOpenGraph  = "{"
	punctCloseGraph = "}"
	punctOpenTuple  = "("
	punctCloseTuple = ")"
)

func decoderErr(tok LexerToken, err error) error {
	return fmt.Errorf("%s: %w", tok.errPrefix(), err)
}

func decoderErrf(tok LexerToken, str string, args ...interface{}) error {
	return decoderErr(tok, fmt.Errorf(str, args...))
}

func isPunct(tok LexerToken, val string) bool {
	return tok.Kind == LexerTokenKindPunctuation && tok.Value == val
}

func isTerm(tok LexerToken) bool {
	return isPunct(tok, punctTerm)
}

// decoder is currently only really used to namespace functions related to
// decoding Graphs. It may later have actual fields added to it, such as for
// options passed by the caller.
type decoder struct{}

// returned boolean value indicates if the token following the single token is a
// term. If a term followed the first token then it is not included in the
// returned leftover tokens.
//
// if termed is false then leftover tokens cannot be empty.
func (d *decoder) parseSingleValue(
	toks []LexerToken,
) (
	Value, []LexerToken, bool, error,
) {

	tok, rest := toks[0], toks[1:]

	if len(rest) == 0 {
		return ZeroValue, nil, false, decoderErrf(tok, "cannot be final token, possibly missing %q", punctTerm)
	}

	termed := isTerm(rest[0])

	if termed {
		rest = rest[1:]
	}

	switch tok.Kind {

	case LexerTokenKindName:
		return Value{Name: &tok.Value, LexerToken: &tok}, rest, termed, nil

	case LexerTokenKindNumber:

		i, err := strconv.ParseInt(tok.Value, 10, 64)

		if err != nil {
			return ZeroValue, nil, false, decoderErrf(tok, "parsing %q as integer: %w", tok.Value, err)
		}

		return Value{Number: &i, LexerToken: &tok}, rest, termed, nil

	case LexerTokenKindPunctuation:
		return ZeroValue, nil, false, decoderErrf(tok, "expected value, found punctuation %q", tok.Value)

	default:
		panic(fmt.Sprintf("unexpected token kind %q", tok.Kind))
	}
}

func (d *decoder) parseOpenEdge(
	toks []LexerToken,
) (
	*graph.OpenEdge[Value], []LexerToken, error,
) {

	if isPunct(toks[0], punctOpenTuple) {
		return d.parseTuple(toks)
	}

	var (
		val    Value
		termed bool
		err    error
	)

	switch {

	case isPunct(toks[0], punctOpenGraph):
		val, toks, termed, err = d.parseGraphValue(toks, true)

	default:
		val, toks, termed, err = d.parseSingleValue(toks)
	}

	if err != nil {
		return nil, nil, err

	}

	if termed {
		return graph.ValueOut[Value](val, ZeroValue), toks, nil
	}

	opTok, toks := toks[0], toks[1:]

	if !isPunct(opTok, punctOp) {
		return nil, nil, decoderErrf(opTok, "must be %q or %q", punctOp, punctTerm)
	}

	if len(toks) == 0 {
		return nil, nil, decoderErrf(opTok, "%q cannot terminate an edge declaration", punctOp)
	}

	oe, toks, err := d.parseOpenEdge(toks)

	if err != nil {
		return nil, nil, err
	}

	oe = graph.TupleOut[Value]([]*graph.OpenEdge[Value]{oe}, val)

	return oe, toks, nil
}

func (d *decoder) parseTuple(
	toks []LexerToken,
) (
	*graph.OpenEdge[Value], []LexerToken, error,
) {

	openTok, toks := toks[0], toks[1:]

	var edges []*graph.OpenEdge[Value]

	for {

		if len(toks) == 0 {
			return nil, nil, decoderErrf(openTok, "no matching %q", punctCloseTuple)

		} else if isPunct(toks[0], punctCloseTuple) {
			toks = toks[1:]
			break
		}

		var (
			oe  *graph.OpenEdge[Value]
			err error
		)

		oe, toks, err = d.parseOpenEdge(toks)

		if err != nil {
			return nil, nil, err
		}

		edges = append(edges, oe)
	}

	// this is a quirk of the syntax, _technically_ a tuple doesn't need a
	// term after it, since it can't be used as an edge value, and so
	// nothing can come after it in the chain.
	if len(toks) > 0 && isTerm(toks[0]) {
		toks = toks[1:]
	}

	return graph.TupleOut[Value](edges, ZeroValue), toks, nil
}

// returned boolean value indicates if the token following the graph is a term.
// If a term followed the first token then it is not included in the returned
// leftover tokens.
//
// if termed is false then leftover tokens cannot be empty.
func (d *decoder) parseGraphValue(
	toks []LexerToken, expectWrappers bool,
) (
	Value, []LexerToken, bool, error,
) {

	var openTok LexerToken

	if expectWrappers {
		openTok, toks = toks[0], toks[1:]
	}

	g := new(graph.Graph[Value])

	for {

		if len(toks) == 0 {

			if !expectWrappers {
				break
			}

			return ZeroValue, nil, false, decoderErrf(openTok, "no matching %q", punctCloseGraph)

		} else if closingTok := toks[0]; isPunct(closingTok, punctCloseGraph) {

			if !expectWrappers {
				return ZeroValue, nil, false, decoderErrf(closingTok, "unexpected %q", punctCloseGraph)
			}

			toks = toks[1:]

			if len(toks) == 0 {
				return ZeroValue, nil, false, decoderErrf(closingTok, "cannot be final token, possibly missing %q", punctTerm)
			}

			break
		}

		var err error

		if g, toks, err = d.parseValIn(g, toks); err != nil {
			return ZeroValue, nil, false, err
		}
	}

	val := Value{Graph: g}

	if !expectWrappers {
		return val, toks, true, nil
	}

	val.LexerToken = &openTok

	termed := isTerm(toks[0])

	if termed {
		toks = toks[1:]
	}

	return val, toks, termed, nil
}

func (d *decoder) parseValIn(into *graph.Graph[Value], toks []LexerToken) (*graph.Graph[Value], []LexerToken, error) {

	if len(toks) == 0 {
		return into, nil, nil

	} else if len(toks) < 3 {
		return nil, nil, decoderErrf(toks[0], `must be of the form "<name> = ..."`)
	}

	dst := toks[0]
	eq := toks[1]
	toks = toks[2:]

	if dst.Kind != LexerTokenKindName {
		return nil, nil, decoderErrf(dst, "must be a name")

	} else if !isPunct(eq, punctAssign) {
		return nil, nil, decoderErrf(eq, "must be %q", punctAssign)
	}

	oe, toks, err := d.parseOpenEdge(toks)

	if err != nil {
		return nil, nil, err
	}

	dstVal := Value{Name: &dst.Value, LexerToken: &dst}

	return into.AddValueIn(oe, dstVal), toks, nil
}

func (d *decoder) decode(lexer Lexer) (*graph.Graph[Value], error) {

	var toks []LexerToken

	for {

		tok, err := lexer.Next()

		if errors.Is(err, io.EOF) {
			break

		} else if err != nil {
			return nil, fmt.Errorf("reading next token: %w", err)
		}

		toks = append(toks, tok)
	}

	val, _, _, err := d.parseGraphValue(toks, false)

	if err != nil {
		return nil, err
	}

	return val.Graph, nil
}

// DecodeLexer reads lexigraphical tokens from the given Lexer and uses them to
// construct a Graph according to the rules of the gg file format. DecodeLexer
// will only return an error if there is a non-EOF file returned from the Lexer,
// or the tokens read cannot be used to construct a valid Graph.
func DecodeLexer(lexer Lexer) (*graph.Graph[Value], error) {
	decoder := &decoder{}
	return decoder.decode(lexer)
}
