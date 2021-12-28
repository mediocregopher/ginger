package gg

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecoder(t *testing.T) {

	i := func(i int64) Value {
		return Value{Number: &i}
	}

	n := func(n string) Value {
		return Value{Name: &n}
	}

	tests := []struct {
		in  string
		exp *Graph
	}{
		{
			in:  "",
			exp: ZeroGraph,
		},
		{
			in:  "out = 1;",
			exp: ZeroGraph.AddValueIn(ValueOut(i(1), ZeroValue), n("out")),
		},
		{
			in:  "out = incr < 1;",
			exp: ZeroGraph.AddValueIn(ValueOut(i(1), n("incr")), n("out")),
		},
		{
			in: "out = a < b < 1;",
			exp: ZeroGraph.AddValueIn(
				TupleOut(
					[]OpenEdge{ValueOut(i(1), n("b"))},
					n("a"),
				),
				n("out"),
			),
		},
		{
			in: "out = a < b < (1; c < 2; d < e < 3;);",
			exp: ZeroGraph.AddValueIn(
				TupleOut(
					[]OpenEdge{TupleOut(
						[]OpenEdge{
							ValueOut(i(1), ZeroValue),
							ValueOut(i(2), n("c")),
							TupleOut(
								[]OpenEdge{ValueOut(i(3), n("e"))},
								n("d"),
							),
						},
						n("b"),
					)},
					n("a"),
				),
				n("out"),
			),
		},
		{
			in: "out = a < b < (1; c < (d < 2; 3;); );",
			exp: ZeroGraph.AddValueIn(
				TupleOut(
					[]OpenEdge{TupleOut(
						[]OpenEdge{
							ValueOut(i(1), ZeroValue),
							TupleOut(
								[]OpenEdge{
									ValueOut(i(2), n("d")),
									ValueOut(i(3), ZeroValue),
								},
								n("c"),
							),
						},
						n("b"),
					)},
					n("a"),
				),
				n("out"),
			),
		},
		{
			in: "out = { a = 1; b = c < d < 2; };",
			exp: ZeroGraph.AddValueIn(
				ValueOut(
					Value{Graph: ZeroGraph.
						AddValueIn(ValueOut(i(1), ZeroValue), n("a")).
						AddValueIn(
							TupleOut(
								[]OpenEdge{
									ValueOut(i(2), n("d")),
								},
								n("c"),
							),
							n("b"),
						),
					},
					ZeroValue,
				),
				n("out"),
			),
		},
		{
			in: "out = a < { b = 1; } < 2;",
			exp: ZeroGraph.AddValueIn(
				TupleOut(
					[]OpenEdge{
						ValueOut(
							i(2),
							Value{Graph: ZeroGraph.
								AddValueIn(ValueOut(i(1), ZeroValue), n("b")),
							},
						),
					},
					n("a"),
				),
				n("out"),
			),
		},
		{
			in: "a = 1; b = 2;",
			exp: ZeroGraph.
				AddValueIn(ValueOut(i(1), ZeroValue), n("a")).
				AddValueIn(ValueOut(i(2), ZeroValue), n("b")),
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {

			r := &mockReader{body: []byte(test.in)}
			lexer := NewLexer(r)

			got, err := DecodeLexer(lexer)
			assert.NoError(t, err)
			assert.True(t, Equal(got, test.exp), "\nexp:%v\ngot:%v", test.exp, got)

		})
	}
}
