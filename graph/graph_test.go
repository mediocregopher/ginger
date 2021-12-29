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
		zeroGraph = new(Graph[S])
	)

	tests := []struct {
		a, b *Graph[S]
		exp  bool
	}{
		{
			a:   zeroGraph,
			b:   zeroGraph,
			exp: true,
		},
		{
			a:   zeroGraph,
			b:   zeroGraph.AddValueIn(ValueOut[S]("in", "incr"), "out"),
			exp: false,
		},
		{
			a:   zeroGraph.AddValueIn(ValueOut[S]("in", "incr"), "out"),
			b:   zeroGraph.AddValueIn(ValueOut[S]("in", "incr"), "out"),
			exp: true,
		},
		{
			a: zeroGraph.AddValueIn(ValueOut[S]("in", "incr"), "out"),
			b: zeroGraph.AddValueIn(TupleOut[S]([]OpenEdge[S]{
				ValueOut[S]("in", "ident"),
				ValueOut[S]("1", "ident"),
			}, "add"), "out"),
			exp: false,
		},
		{
			// tuples are different order
			a: zeroGraph.AddValueIn(TupleOut[S]([]OpenEdge[S]{
				ValueOut[S]("1", "ident"),
				ValueOut[S]("in", "ident"),
			}, "add"), "out"),
			b: zeroGraph.AddValueIn(TupleOut[S]([]OpenEdge[S]{
				ValueOut[S]("in", "ident"),
				ValueOut[S]("1", "ident"),
			}, "add"), "out"),
			exp: false,
		},
		{
			// tuple with no edge value and just a single input edge should be
			// equivalent to just that edge.
			a: zeroGraph.AddValueIn(TupleOut[S]([]OpenEdge[S]{
				ValueOut[S]("1", "ident"),
			}, zeroValue), "out"),
			b:   zeroGraph.AddValueIn(ValueOut[S]("1", "ident"), "out"),
			exp: true,
		},
		{
			// tuple with an edge value and just a single input edge that has no
			// edgeVal should be equivalent to just that edge with the tuple's
			// edge value.
			a: zeroGraph.AddValueIn(TupleOut[S]([]OpenEdge[S]{
				ValueOut[S]("1", zeroValue),
			}, "ident"), "out"),
			b:   zeroGraph.AddValueIn(ValueOut[S]("1", "ident"), "out"),
			exp: true,
		},
		{
			a: zeroGraph.
				AddValueIn(ValueOut[S]("in", "incr"), "out").
				AddValueIn(ValueOut[S]("in2", "incr2"), "out2"),
			b: zeroGraph.
				AddValueIn(ValueOut[S]("in", "incr"), "out"),
			exp: false,
		},
		{
			a: zeroGraph.
				AddValueIn(ValueOut[S]("in", "incr"), "out").
				AddValueIn(ValueOut[S]("in2", "incr2"), "out2"),
			b: zeroGraph.
				AddValueIn(ValueOut[S]("in", "incr"), "out").
				AddValueIn(ValueOut[S]("in2", "incr2"), "out2"),
			exp: true,
		},
		{
			// order of value ins shouldn't matter
			a: zeroGraph.
				AddValueIn(ValueOut[S]("in", "incr"), "out").
				AddValueIn(ValueOut[S]("in2", "incr2"), "out2"),
			b: zeroGraph.
				AddValueIn(ValueOut[S]("in2", "incr2"), "out2").
				AddValueIn(ValueOut[S]("in", "incr"), "out"),
			exp: true,
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert.Equal(t, test.exp, test.a.Equal(test.b))
		})
	}
}
