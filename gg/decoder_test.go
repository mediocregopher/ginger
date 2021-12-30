package gg

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mediocregopher/ginger/graph"
)

func TestDecoder(t *testing.T) {

	zeroGraph := new(Graph)

	i := func(i int64) Value {
		return Value{Number: &i}
	}

	n := func(n string) Value {
		return Value{Name: &n}
	}

	vOut := func(edgeVal, val Value) *OpenEdge {
		return graph.ValueOut(edgeVal, val)
	}

	tOut := func(edgeVal Value, ins ...*OpenEdge) *OpenEdge {
		return graph.TupleOut(edgeVal, ins...)
	}

	tests := []struct {
		in  string
		exp *Graph
	}{
		{
			in:  "",
			exp: zeroGraph,
		},
		{
			in:  "out = 1;",
			exp: zeroGraph.AddValueIn(n("out"), vOut(ZeroValue, i(1))),
		},
		{
			in:  "out = incr < 1;",
			exp: zeroGraph.AddValueIn(n("out"), vOut(n("incr"), i(1))),
		},
		{
			in: "out = a < b < 1;",
			exp: zeroGraph.AddValueIn(
				n("out"),
				tOut(
					n("a"),
					vOut(n("b"),
						i(1)),
				),
			),
		},
		{
			in: "out = a < b < (1; c < 2; d < e < 3;);",
			exp: zeroGraph.AddValueIn(
				n("out"),
				tOut(
					n("a"),
					tOut(
						n("b"),
						vOut(ZeroValue, i(1)),
						vOut(n("c"), i(2)),
						tOut(
							n("d"),
							vOut(n("e"), i(3)),
						),
					),
				),
			),
		},
		{
			in: "out = a < b < (1; c < (d < 2; 3;); );",
			exp: zeroGraph.AddValueIn(
				n("out"),
				tOut(
					n("a"),
					tOut(
						n("b"),
						vOut(ZeroValue, i(1)),
						tOut(
							n("c"),
							vOut(n("d"), i(2)),
							vOut(ZeroValue, i(3)),
						),
					),
				),
			),
		},
		{
			in: "out = { a = 1; b = c < d < 2; };",
			exp: zeroGraph.AddValueIn(
				n("out"),
				vOut(
					ZeroValue,
					Value{Graph: zeroGraph.
						AddValueIn(n("a"), vOut(ZeroValue, i(1))).
						AddValueIn(
							n("b"),
							tOut(
								n("c"),
								vOut(n("d"), i(2)),
							),
						),
					},
				),
			),
		},
		{
			in: "out = a < { b = 1; } < 2;",
			exp: zeroGraph.AddValueIn(
				n("out"),
				tOut(
					n("a"),
					vOut(
						Value{Graph: zeroGraph.
							AddValueIn(n("b"), vOut(ZeroValue, i(1))),
						},
						i(2),
					),
				),
			),
		},
		{
			in: "a = 1; b = 2;",
			exp: zeroGraph.
				AddValueIn(n("a"), vOut(ZeroValue, i(1))).
				AddValueIn(n("b"), vOut(ZeroValue, i(2))),
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
