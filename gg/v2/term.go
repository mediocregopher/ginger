package gg

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	"github.com/mediocregopher/ginger/graph"
)

var (
	errNoMatch = errors.New("not found")
)

type stringerFn func() string

func (fn stringerFn) String() string {
	return fn()
}

type stringerStr string

func (str stringerStr) String() string {
	return string(str)
}

type term[T locatable] struct {
	name     fmt.Stringer
	decodeFn func(d *Decoder) (T, error)
}

func (t term[T]) String() string {
	return t.name.String()
}

func firstOf[T locatable](terms ...*term[T]) *term[T] {
	if len(terms) < 2 {
		panic("firstOfTerms requires at least 2 terms")
	}

	return &term[T]{
		name: stringerFn(func() string {
			descrs := make([]string, len(terms))
			for i := range terms {
				descrs[i] = terms[i].String()
			}
			return strings.Join(descrs, " or ")
		}),
		decodeFn: func(d *Decoder) (T, error) {
			var zero T
			for _, t := range terms {
				v, err := t.decodeFn(d)
				if errors.Is(err, errNoMatch) {
					continue
				} else if err != nil {
					return zero, err
				}

				return v, nil
			}

			return zero, errNoMatch
		},
	}
}

func seq[Ta, Tb, Tc locatable](
	name fmt.Stringer,
	termA *term[Ta],
	termB *term[Tb],
	fn func(Ta, Tb) (Tc, error),
) *term[Tc] {
	return &term[Tc]{
		name: name,
		decodeFn: func(d *Decoder) (Tc, error) {
			var zero Tc

			va, err := termA.decodeFn(d)
			if err != nil {
				return zero, err
			}

			vb, err := termB.decodeFn(d)
			if errors.Is(err, errNoMatch) {
				return zero, d.nextLoc().errf("expected %v", termB)
			} else if err != nil {
				return zero, err
			}

			vc, err := fn(va, vb)
			if err != nil {
				return zero, err
			}

			return vc, nil
		},
	}
}

func matchAndSkip[Ta, Tb locatable](
	termA *term[Ta], termB *term[Tb],
) *term[Tb] {
	return seq(termA, termA, termB, func(_ Ta, b Tb) (Tb, error) {
		return b, nil
	})
}

func oneOrMore[T locatable](t *term[T]) *term[locatableSlice[T]] {
	return &term[locatableSlice[T]]{
		name: stringerFn(func() string {
			return fmt.Sprintf("one or more %v", t)
		}),
		decodeFn: func(d *Decoder) (locatableSlice[T], error) {
			var vv []T
			for {
				v, err := t.decodeFn(d)
				if errors.Is(err, errNoMatch) {
					break
				} else if err != nil {
					return nil, err
				}

				vv = append(vv, v)
			}

			if len(vv) == 0 {
				return nil, errNoMatch
			}

			return vv, nil
		},
	}
}

func zeroOrMore[T locatable](t *term[T]) *term[locatableSlice[T]] {
	return &term[locatableSlice[T]]{
		name: stringerFn(func() string {
			return fmt.Sprintf("zero or more %v", t)
		}),
		decodeFn: func(d *Decoder) (locatableSlice[T], error) {
			var vv []T
			for {
				v, err := t.decodeFn(d)
				if errors.Is(err, errNoMatch) {
					break
				} else if err != nil {
					return nil, err
				}

				vv = append(vv, v)
			}

			return vv, nil
		},
	}
}

func mapTerm[Ta locatable, Tb locatable](
	name fmt.Stringer, t *term[Ta], fn func(Ta) Tb,
) *term[Tb] {
	return &term[Tb]{
		name: name,
		decodeFn: func(d *Decoder) (Tb, error) {
			var zero Tb
			va, err := t.decodeFn(d)
			if err != nil {
				return zero, err
			}
			return fn(va), nil
		},
	}
}

func runePredTerm(
	name fmt.Stringer, pred func(rune) bool,
) *term[locatableRune] {
	return &term[locatableRune]{
		name: name,
		decodeFn: func(d *Decoder) (locatableRune, error) {
			lr, err := d.readRune()
			if errors.Is(err, io.EOF) {
				return locatableRune{}, errNoMatch
			} else if err != nil {
				return locatableRune{}, err
			}

			if !pred(lr.r) {
				d.unreadRune(lr)
				return locatableRune{}, errNoMatch
			}

			return lr, nil
		},
	}
}

func runeTerm(r rune) *term[locatableRune] {
	return runePredTerm(
		stringerStr(fmt.Sprintf("'%c'", r)),
		func(r2 rune) bool { return r2 == r },
	)
}

func locatableRunesToString(rr locatableSlice[locatableRune]) string {
	str := make([]rune, len(rr))
	for i := range rr {
		str[i] = rr[i].r
	}
	return string(str)
}

func runesToStringTerm(
	t *term[locatableSlice[locatableRune]],
) *term[locatableString] {
	return mapTerm(
		t, t, func(rr locatableSlice[locatableRune]) locatableString {
			return locatableString{rr.locate(), locatableRunesToString(rr)}
		},
	)
}

var (
	digitTerm = runePredTerm(
		stringerStr("digit"),
		func(r rune) bool { return '0' <= r && r <= '9' },
	)

	positiveNumberTerm = runesToStringTerm(oneOrMore(digitTerm))

	negativeNumberTerm = seq(
		stringerStr("negative-number"),
		runeTerm('-'),
		positiveNumberTerm,
		func(neg locatableRune, posNum locatableString) (locatableString, error) {
			return locatableString{
				neg.locate(), string(neg.r) + posNum.str,
			}, nil
		},
	)

	numberTerm = mapTerm(
		stringerStr("number"),
		firstOf(negativeNumberTerm, positiveNumberTerm),
		func(str locatableString) Value {
			i, err := strconv.ParseInt(str.str, 10, 64)
			if err != nil {
				panic(fmt.Errorf("parsing %q as int: %w", str, err))
			}

			return Value{Number: &i, Location: str.locate()}
		},
	)
)

var (
	letterTerm = runePredTerm(
		stringerStr("letter"),
		func(r rune) bool {
			return unicode.In(r, unicode.Letter, unicode.Mark)
		},
	)

	letterTailTerm = zeroOrMore(firstOf(letterTerm, digitTerm))

	nameTerm = seq(
		stringerStr("name"),
		letterTerm,
		letterTailTerm,
		func(head locatableRune, tail locatableSlice[locatableRune]) (Value, error) {
			name := string(head.r) + locatableRunesToString(tail)
			return Value{Name: &name, Location: head.locate()}, nil
		},
	)
)

var graphTerm = func() *term[Value] {
	type graphState struct {
		Location // location of last place graphState was updated
		g        *Graph
		oe       *OpenEdge
	}

	var (
		// pre-define these, and then fill in the pointers after, in order to
		// deal with recursive dependencies between them.
		graphTerm             = new(term[Value])
		graphTailTerm         = new(term[graphState])
		graphOpenEdgeTerm     = new(term[graphState])
		graphOpenEdgeTailTerm = new(term[graphState])
		valueTerm             = new(term[Value])

		rightCurlyBrace = runeTerm('}')
		graphEndTerm    = mapTerm(
			rightCurlyBrace,
			rightCurlyBrace, func(lr locatableRune) graphState {
				// if '}', then map that to an empty state. This acts as a
				// sentinel value to indicate "end of graph".
				return graphState{Location: lr.locate()}
			},
		)
	)

	*graphTerm = *seq(
		stringerStr("graph"),
		runeTerm('{'),
		graphTailTerm,
		func(lr locatableRune, gs graphState) (Value, error) {
			if gs.g == nil {
				gs.g = new(Graph)
			}

			return Value{Graph: gs.g, Location: lr.locate()}, nil
		},
	)

	*graphTailTerm = *firstOf(
		graphEndTerm,
		seq(
			nameTerm,
			nameTerm,
			matchAndSkip(runeTerm('='), graphOpenEdgeTailTerm),
			func(name Value, gs graphState) (graphState, error) {
				if gs.g == nil {
					gs.g = new(Graph)
				}

				gs.g = gs.g.AddValueIn(name, gs.oe)
				gs.oe = nil
				gs.Location = name.locate()
				return gs, nil
			},
		),
	)

	*graphOpenEdgeTerm = *firstOf(
		graphEndTerm,
		matchAndSkip(runeTerm(';'), graphTailTerm),
		matchAndSkip(runeTerm('<'), graphOpenEdgeTailTerm),
	)

	*graphOpenEdgeTailTerm = *seq(
		valueTerm,
		valueTerm,
		graphOpenEdgeTerm,
		func(val Value, gs graphState) (graphState, error) {
			if gs.oe == nil {
				gs.oe = graph.ValueOut(None, val)
			} else if !gs.oe.EdgeValue().Valid {
				gs.oe = gs.oe.WithEdgeValue(Some(val))
			} else {
				gs.oe = graph.TupleOut(Some(val), gs.oe)
			}

			gs.Location = val.locate()
			return gs, nil
		},
	)

	*valueTerm = *firstOf(nameTerm, numberTerm, graphTerm)

	return graphTerm
}()
