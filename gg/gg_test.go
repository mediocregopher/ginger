package gg

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEqual(t *testing.T) {

	i := func(i int64) Value {
		return Value{Number: &i}
	}

	n := func(n string) Value {
		return Value{Name: &n}
	}

	tests := []struct {
		a, b *Graph
		exp  bool
	}{
		{
			a:   ZeroGraph,
			b:   ZeroGraph,
			exp: true,
		},
		{
			a:   ZeroGraph,
			b:   ZeroGraph.AddValueIn(ValueOut(n("in"), n("incr")), n("out")),
			exp: false,
		},
		{
			a:   ZeroGraph.AddValueIn(ValueOut(n("in"), n("incr")), n("out")),
			b:   ZeroGraph.AddValueIn(ValueOut(n("in"), n("incr")), n("out")),
			exp: true,
		},
		{
			a: ZeroGraph.AddValueIn(ValueOut(n("in"), n("incr")), n("out")),
			b: ZeroGraph.AddValueIn(TupleOut([]OpenEdge{
				ValueOut(n("in"), n("ident")),
				ValueOut(i(1), n("ident")),
			}, n("add")), n("out")),
			exp: false,
		},
		{
			// tuples are different order
			a: ZeroGraph.AddValueIn(TupleOut([]OpenEdge{
				ValueOut(i(1), n("ident")),
				ValueOut(n("in"), n("ident")),
			}, n("add")), n("out")),
			b: ZeroGraph.AddValueIn(TupleOut([]OpenEdge{
				ValueOut(n("in"), n("ident")),
				ValueOut(i(1), n("ident")),
			}, n("add")), n("out")),
			exp: false,
		},
		{
			a: ZeroGraph.
				AddValueIn(ValueOut(n("in"), n("incr")), n("out")).
				AddValueIn(ValueOut(n("in2"), n("incr2")), n("out2")),
			b: ZeroGraph.
				AddValueIn(ValueOut(n("in"), n("incr")), n("out")),
			exp: false,
		},
		{
			a: ZeroGraph.
				AddValueIn(ValueOut(n("in"), n("incr")), n("out")).
				AddValueIn(ValueOut(n("in2"), n("incr2")), n("out2")),
			b: ZeroGraph.
				AddValueIn(ValueOut(n("in"), n("incr")), n("out")).
				AddValueIn(ValueOut(n("in2"), n("incr2")), n("out2")),
			exp: true,
		},
		{
			// order of value ins shouldn't matter
			a: ZeroGraph.
				AddValueIn(ValueOut(n("in"), n("incr")), n("out")).
				AddValueIn(ValueOut(n("in2"), n("incr2")), n("out2")),
			b: ZeroGraph.
				AddValueIn(ValueOut(n("in2"), n("incr2")), n("out2")).
				AddValueIn(ValueOut(n("in"), n("incr")), n("out")),
			exp: true,
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert.Equal(t, test.exp, Equal(test.a, test.b))
		})
	}
}
