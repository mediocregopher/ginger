package graph

import (
	"fmt"
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

type S string

func (s S) Equal(s2 Value) bool { return s == s2.(S) }

func (s S) String() string { return string(s) }

type I int

func (i I) Equal(i2 Value) bool { return i == i2.(I) }

func (i I) String() string { return strconv.Itoa(int(i)) }

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

type mapReduceTestEdge struct {
	name string
	fn func([]int) int
	done bool
}

func (e *mapReduceTestEdge) Equal(e2i Value) bool {

	e2, _ := e2i.(*mapReduceTestEdge)

	if e == nil || e2 == nil {
		return e == e2
	}

	return e.name == e2.name
}

func (e *mapReduceTestEdge) String() string {
	return e.name
}

func (e *mapReduceTestEdge) do(ii []int) int {

	if e.done {
		panic(fmt.Sprintf("%q already done", e.name))
	}

	e.done = true

	return e.fn(ii)
}

func TestMapReduce(t *testing.T) {

	type (
		Va = I
		Vb = int
		Ea = *mapReduceTestEdge
		edge = OpenEdge[Ea, Va]
	)

	var (
		zeroVb Vb
	)

	vOut := func(edge Ea, val Va) *edge {
		return ValueOut[Ea, Va](edge, val)
	}

	tOut := func(edge Ea, ins ...*edge) *edge {
		return TupleOut[Ea, Va](edge, ins...)
	}

	add := func() *mapReduceTestEdge{
		return &mapReduceTestEdge{
			name: "add",
			fn: func(ii []int) int {
				var n int
				for _, i := range ii {
					n += i
				}
				return n
			},
		}
	}

	mul := func() *mapReduceTestEdge{
		return &mapReduceTestEdge{
			name: "mul",
			fn: func(ii []int) int {
				n := 1
				for _, i := range ii {
					n *= i
				}
				return n
			},
		}
	}

	mapVal := func(valA Va) (Vb, error) {
		return Vb(valA * 10), nil
	}

	reduceEdge := func(edgeA Ea, valBs []Vb) (Vb, error) {

		if edgeA == nil {

			if len(valBs) == 1 {
				return valBs[0], nil
			}

			return zeroVb, errors.New("tuple edge must have edge value")
		}

		return edgeA.do(valBs), nil
	}

	tests := []struct {
		in *edge
		exp int
	}{
		{
			in: vOut(nil, 1),
			exp: 10,
		},
		{
			in: vOut(add(), 1),
			exp: 10,
		},
		{
			in: tOut(
				add(),
				vOut(nil, 1),
				vOut(add(), 2),
				vOut(mul(), 3),
			),
			exp: 60,
		},
		{
			// duplicate edges and values getting used twice, each should only
			// get eval'd once
			in: tOut(
				add(),
				tOut(add(), vOut(nil, 1), vOut(nil, 2)),
				tOut(add(), vOut(nil, 1), vOut(nil, 2)),
				tOut(add(), vOut(nil, 3), vOut(nil, 3)),
			),
			exp: 120,
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got, err := MapReduce(test.in, mapVal, reduceEdge)
			assert.NoError(t, err)
			assert.Equal(t, test.exp, got)
		})
	}
}
