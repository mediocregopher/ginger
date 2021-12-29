package gg

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mediocregopher/ginger/graph"
)

func TestDecoder(t *testing.T) {

	zeroGraph := new(graph.Graph[Value])

	i := func(i int64) Value {
		return Value{Number: &i}
	}

	n := func(n string) Value {
		return Value{Name: &n}
	}

	vOut := func(val, edgeVal Value) graph.OpenEdge[Value] {
		return graph.ValueOut(val, edgeVal)
	}

	tOut := func(ins []graph.OpenEdge[Value], edgeVal Value) graph.OpenEdge[Value] {
		return graph.TupleOut(ins, edgeVal)
	}

	type openEdge = graph.OpenEdge[Value]

	tests := []struct {
		in  string
		exp *graph.Graph[Value]
	}{
		{
			in:  "",
			exp: zeroGraph,
		},
		{
			in:  "out = 1;",
			exp: zeroGraph.AddValueIn(vOut(i(1), ZeroValue), n("out")),
		},
		{
			in:  "out = incr < 1;",
			exp: zeroGraph.AddValueIn(vOut(i(1), n("incr")), n("out")),
		},
		{
			in: "out = a < b < 1;",
			exp: zeroGraph.AddValueIn(
				tOut(
					[]openEdge{vOut(i(1), n("b"))},
					n("a"),
				),
				n("out"),
			),
		},
		{
			in: "out = a < b < (1; c < 2; d < e < 3;);",
			exp: zeroGraph.AddValueIn(
				tOut(
					[]openEdge{tOut(
						[]openEdge{
							vOut(i(1), ZeroValue),
							vOut(i(2), n("c")),
							tOut(
								[]openEdge{vOut(i(3), n("e"))},
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
			exp: zeroGraph.AddValueIn(
				tOut(
					[]openEdge{tOut(
						[]openEdge{
							vOut(i(1), ZeroValue),
							tOut(
								[]openEdge{
									vOut(i(2), n("d")),
									vOut(i(3), ZeroValue),
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
			exp: zeroGraph.AddValueIn(
				vOut(
					Value{Graph: zeroGraph.
						AddValueIn(vOut(i(1), ZeroValue), n("a")).
						AddValueIn(
							tOut(
								[]openEdge{
									vOut(i(2), n("d")),
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
			exp: zeroGraph.AddValueIn(
				tOut(
					[]openEdge{
						vOut(
							i(2),
							Value{Graph: zeroGraph.
								AddValueIn(vOut(i(1), ZeroValue), n("b")),
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
			exp: zeroGraph.
				AddValueIn(vOut(i(1), ZeroValue), n("a")).
				AddValueIn(vOut(i(2), ZeroValue), n("b")),
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {

			r := &mockReader{body: []byte(test.in)}
			lexer := NewLexer(r)

			got, err := DecodeLexer(lexer)
			assert.NoError(t, err)
			assert.True(t, got.Equal(test.exp), "\nexp:%v\ngot:%v", test.exp, got)

		})
	}
}
