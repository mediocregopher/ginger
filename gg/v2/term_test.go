package gg

import (
	"bytes"
	"io"
	"strconv"
	"testing"

	"github.com/mediocregopher/ginger/graph"
	"github.com/stretchr/testify/assert"
)

func decoderLeftover(d *Decoder) string {
	unread := make([]rune, len(d.unread))
	for i := range unread {
		unread[i] = d.unread[i].r
	}

	rest, err := io.ReadAll(d.br)
	if err != nil {
		panic(err)
	}
	return string(unread) + string(rest)
}

func TestTermDecoding(t *testing.T) {
	type test struct {
		in       string
		exp      Value
		expErr   string
		leftover string
	}

	runTests := func(
		t *testing.T, name string, term *term[Value], tests []test,
	) {
		t.Run(name, func(t *testing.T) {
			for i, test := range tests {
				t.Run(strconv.Itoa(i), func(t *testing.T) {
					dec := NewDecoder(bytes.NewBufferString(test.in))
					got, err := term.decodeFn(dec)
					if test.expErr != "" {
						assert.Error(t, err)
						assert.Equal(t, test.expErr, err.Error())
					} else if assert.NoError(t, err) {
						assert.True(t,
							test.exp.Equal(got),
							"\nexp:%v\ngot:%v", test.exp, got,
						)
						assert.Equal(t, test.leftover, decoderLeftover(dec))
					}
				})
			}
		})
	}

	expNum := func(row, col int, n int64) Value {
		return Value{Number: &n, Location: Location{row, col}}
	}

	runTests(t, "number", numberTerm, []test{
		{in: `0`, exp: expNum(1, 1, 0)},
		{in: `100`, exp: expNum(1, 1, 100)},
		{in: `-100`, exp: expNum(1, 1, -100)},
		{in: `0foo`, exp: expNum(1, 1, 0), leftover: "foo"},
		{in: `100foo`, exp: expNum(1, 1, 100), leftover: "foo"},
	})

	expName := func(row, col int, name string) Value {
		return Value{Name: &name, Location: Location{row, col}}
	}

	expGraph := func(row, col int, g *Graph) Value {
		return Value{Graph: g, Location: Location{row, col}}
	}

	runTests(t, "name", nameTerm, []test{
		{in: `a`, exp: expName(1, 1, "a")},
		{in: `ab`, exp: expName(1, 1, "ab")},
		{in: `ab2c`, exp: expName(1, 1, "ab2c")},
	})

	runTests(t, "graph", graphTerm, []test{
		{in: `{}`, exp: expGraph(1, 1, new(Graph))},
		{in: `{`, expErr: `1:2: expected '}' or name`},
		{in: `{a}`, expErr: `1:3: expected '='`},
		{in: `{a=}`, expErr: `1:4: expected name or number or graph`},
		{
			in: `{foo=a}`,
			exp: expGraph(
				1, 1, new(Graph).
					AddValueIn(
						expName(2, 1, "foo"),
						graph.ValueOut(None, expName(6, 1, "a")),
					),
			),
		},
		{in: `{1=a}`, expErr: `1:2: expected '}' or name`},
		{in: `{foo=a,}`, expErr: `1:7: expected '}' or ';' or '<'`},
		{in: `{foo=a`, expErr: `1:7: expected '}' or ';' or '<'`},
		{
			in: `{foo=a<b}`,
			exp: expGraph(
				1, 1, new(Graph).
					AddValueIn(
						expName(2, 1, "foo"),
						graph.ValueOut(
							Some(expName(6, 1, "a")),
							expName(8, 1, "b"),
						),
					),
			),
		},
		{
			in: `{foo=a<b<c}`,
			exp: expGraph(
				1, 1, new(Graph).
					AddValueIn(
						expName(2, 1, "foo"),
						graph.TupleOut(
							Some(expName(6, 1, "a")),
							graph.ValueOut(
								Some(expName(8, 1, "b")),
								expName(10, 1, "c"),
							),
						),
					),
			),
		},
		{
			in: `{foo=a<b<c<1}`,
			exp: expGraph(
				1, 1, new(Graph).
					AddValueIn(
						expName(2, 1, "foo"),
						graph.TupleOut(
							Some(expName(6, 1, "a")),
							graph.TupleOut(
								Some(expName(8, 1, "b")),
								graph.ValueOut(
									Some(expName(10, 1, "c")),
									expNum(12, 1, 1),
								),
							),
						),
					),
			),
		},
		{
			in: `{foo=a<b;}`,
			exp: expGraph(
				1, 1, new(Graph).
					AddValueIn(
						expName(2, 1, "foo"),
						graph.ValueOut(
							Some(expName(6, 1, "a")),
							expName(8, 1, "b"),
						),
					),
			),
		},
		{
			in: `{foo=a<b;bar=c}`,
			exp: expGraph(
				1, 1, new(Graph).
					AddValueIn(
						expName(2, 1, "foo"),
						graph.ValueOut(
							Some(expName(6, 1, "a")),
							expName(8, 1, "b"),
						),
					).
					AddValueIn(
						expName(10, 1, "bar"),
						graph.ValueOut(None, expName(15, 1, "c")),
					),
			),
		},
		{
			in: `{foo=a<{baz=1};bar=c}`,
			exp: expGraph(
				1, 1, new(Graph).
					AddValueIn(
						expName(2, 1, "foo"),
						graph.ValueOut(
							Some(expName(6, 1, "a")),
							expGraph(8, 1, new(Graph).AddValueIn(
								expName(9, 1, "baz"),
								graph.ValueOut(None, expNum(13, 1, 1)),
							)),
						),
					).
					AddValueIn(
						expName(16, 1, "bar"),
						graph.ValueOut(None, expName(20, 1, "c")),
					),
			),
		},
		{
			in: `{foo={baz=1}<a;bar=c}`,
			exp: expGraph(
				1, 1, new(Graph).
					AddValueIn(
						expName(2, 1, "foo"),
						graph.ValueOut(
							Some(expGraph(8, 1, new(Graph).AddValueIn(
								expName(9, 1, "baz"),
								graph.ValueOut(None, expNum(13, 1, 1)),
							))),
							expName(6, 1, "a"),
						),
					).
					AddValueIn(
						expName(16, 1, "bar"),
						graph.ValueOut(None, expName(20, 1, "c")),
					),
			),
		},
	})
}
