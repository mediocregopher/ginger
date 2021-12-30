package graph

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

type S string

func (s S) Equal(s2 Value) bool { return s == s2.(S) }

func (s S) String() string { return string(s) }

func TestEqual(t *testing.T) {

	var (
		zeroValue S
		zeroGraph = new(Graph[S, S])
	)

	tests := []struct {
		a, b *Graph[S, S]
		exp  bool
	}{
		{
			a:   zeroGraph,
			b:   zeroGraph,
			exp: true,
		},
		{
			a:   zeroGraph,
			b:   zeroGraph.AddValueIn("out", ValueOut[S, S]("incr", "in")),
			exp: false,
		},
		{
			a:   zeroGraph.AddValueIn("out", ValueOut[S, S]("incr", "in")),
			b:   zeroGraph.AddValueIn("out", ValueOut[S, S]("incr", "in")),
			exp: true,
		},
		{
			a: zeroGraph.AddValueIn("out", ValueOut[S, S]("incr", "in")),
			b: zeroGraph.AddValueIn("out", TupleOut[S, S](
				"add",
				ValueOut[S, S]("ident", "in"),
				ValueOut[S, S]("ident", "1"),
			)),
			exp: false,
		},
		{
			// tuples are different order
			a: zeroGraph.AddValueIn("out", TupleOut[S, S](
				"add",
				ValueOut[S, S]("ident", "1"),
				ValueOut[S, S]("ident", "in"),
			)),
			b: zeroGraph.AddValueIn("out", TupleOut[S, S](
				"add",
				ValueOut[S, S]("ident", "in"),
				ValueOut[S, S]("ident", "1"),
			)),
			exp: false,
		},
		{
			// tuple with no edge value and just a single input edge should be
			// equivalent to just that edge.
			a: zeroGraph.AddValueIn("out", TupleOut[S, S](
				zeroValue,
				ValueOut[S, S]("ident", "1"),
			)),
			b:   zeroGraph.AddValueIn("out", ValueOut[S, S]("ident", "1")),
			exp: true,
		},
		{
			// tuple with an edge value and just a single input edge that has no
			// edgeVal should be equivalent to just that edge with the tuple's
			// edge value.
			a: zeroGraph.AddValueIn("out", TupleOut[S, S](
				"ident",
				ValueOut[S, S](zeroValue, "1"),
			)),
			b:   zeroGraph.AddValueIn("out", ValueOut[S, S]("ident", "1")),
			exp: true,
		},
		{
			a: zeroGraph.
				AddValueIn("out", ValueOut[S, S]("incr", "in")).
				AddValueIn("out2", ValueOut[S, S]("incr2", "in2")),
			b: zeroGraph.
				AddValueIn("out", ValueOut[S, S]("incr", "in")),
			exp: false,
		},
		{
			a: zeroGraph.
				AddValueIn("out", ValueOut[S, S]("incr", "in")).
				AddValueIn("out2", ValueOut[S, S]("incr2", "in2")),
			b: zeroGraph.
				AddValueIn("out", ValueOut[S, S]("incr", "in")).
				AddValueIn("out2", ValueOut[S, S]("incr2", "in2")),
			exp: true,
		},
		{
			// order of value ins shouldn't matter
			a: zeroGraph.
				AddValueIn("out", ValueOut[S, S]("incr", "in")).
				AddValueIn("out2", ValueOut[S, S]("incr2", "in2")),
			b: zeroGraph.
				AddValueIn("out2", ValueOut[S, S]("incr2", "in2")).
				AddValueIn("out", ValueOut[S, S]("incr", "in")),
			exp: true,
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert.Equal(t, test.exp, test.a.Equal(test.b))
		})
	}
}
