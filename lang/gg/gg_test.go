package gg

import (
	"fmt"
	"hash"
	. "testing"

	"github.com/stretchr/testify/assert"
)

type idAny struct {
	i interface{}
}

func (i idAny) Identify(h hash.Hash) {
	fmt.Fprintln(h, i)
}

func id(i interface{}) Identifier {
	return idAny{i: i}
}

func edge(val string, from *Vertex) Edge {
	return Edge{Value: id(val), From: from}
}

func value(val string, in ...Edge) *Vertex {
	return &Vertex{
		VertexType: Value,
		Value:      id(val),
		In:         in,
	}
}

func junction(val string, in ...Edge) Edge {
	return Edge{
		From: &Vertex{
			VertexType: Junction,
			In:         in,
		},
		Value: id(val),
	}
}

func assertVertexEqual(t *T, exp, got *Vertex, msgAndArgs ...interface{}) bool {
	var assertInner func(*Vertex, *Vertex, map[*Vertex]bool) bool
	assertInner = func(exp, got *Vertex, m map[*Vertex]bool) bool {
		// if got is already in m then we've already looked at it
		if m[got] {
			return true
		}
		m[got] = true

		assert.Equal(t, exp.VertexType, got.VertexType, msgAndArgs...)
		assert.Equal(t, exp.Value, got.Value, msgAndArgs...)
		if !assert.Len(t, got.In, len(exp.In), msgAndArgs...) {
			return false
		}
		for i := range exp.In {
			assertInner(exp.In[i].From, got.In[i].From, m)
			assert.Equal(t, exp.In[i].Value, got.In[i].Value, msgAndArgs...)
			assert.Equal(t, got, got.In[i].To)
			assert.Contains(t, got.In[i].From.Out, got.In[i])
		}
		return true

	}
	return assertInner(exp, got, map[*Vertex]bool{})
}

type graphTest struct {
	name string
	out  func() *Graph
	exp  []*Vertex
}

func mkTest(name string, out func() *Graph, exp ...*Vertex) graphTest {
	return graphTest{name: name, out: out, exp: exp}
}

func TestGraph(t *T) {
	tests := []graphTest{
		mkTest(
			"values-basic",
			func() *Graph {
				return Null.AddValueIn(ValueOut(id("v0"), id("e0")), id("v1"))
			},
			value("v0"),
			value("v1", edge("e0", value("v0"))),
		),

		mkTest(
			"values-2edges",
			func() *Graph {
				g0 := Null.AddValueIn(ValueOut(id("v0"), id("e0")), id("v2"))
				return g0.AddValueIn(ValueOut(id("v1"), id("e1")), id("v2"))
			},
			value("v0"),
			value("v1"),
			value("v2",
				edge("e0", value("v0")),
				edge("e1", value("v1")),
			),
		),

		mkTest(
			"values-separate",
			func() *Graph {
				g0 := Null.AddValueIn(ValueOut(id("v0"), id("e0")), id("v1"))
				return g0.AddValueIn(ValueOut(id("v2"), id("e2")), id("v3"))
			},
			value("v0"),
			value("v1", edge("e0", value("v0"))),
			value("v2"),
			value("v3", edge("e2", value("v2"))),
		),

		mkTest(
			"values-circular",
			func() *Graph {
				return Null.AddValueIn(ValueOut(id("v0"), id("e")), id("v0"))
			},
			value("v0", edge("e", value("v0"))),
		),

		mkTest(
			"values-circular2",
			func() *Graph {
				g0 := Null.AddValueIn(ValueOut(id("v0"), id("e0")), id("v1"))
				return g0.AddValueIn(ValueOut(id("v1"), id("e1")), id("v0"))
			},
			value("v0", edge("e1", value("v1", edge("e0", value("v0"))))),
			value("v1", edge("e0", value("v0", edge("e1", value("v1"))))),
		),

		mkTest(
			"values-circular3",
			func() *Graph {
				g0 := Null.AddValueIn(ValueOut(id("v0"), id("e0")), id("v1"))
				g1 := g0.AddValueIn(ValueOut(id("v1"), id("e1")), id("v2"))
				return g1.AddValueIn(ValueOut(id("v2"), id("e2")), id("v1"))
			},
			value("v0"),
			value("v1",
				edge("e0", value("v0")),
				edge("e2", value("v2", edge("e1", value("v1")))),
			),
			value("v2", edge("e1", value("v1",
				edge("e0", value("v0")),
				edge("e2", value("v2")),
			))),
		),

		mkTest(
			"junction-basic",
			func() *Graph {
				e0 := ValueOut(id("v0"), id("e0"))
				e1 := ValueOut(id("v1"), id("e1"))
				ej0 := JunctionOut([]HalfEdge{e0, e1}, id("ej0"))
				return Null.AddValueIn(ej0, id("v2"))
			},
			value("v0"), value("v1"),
			value("v2", junction("ej0",
				edge("e0", value("v0")),
				edge("e1", value("v1")),
			)),
		),

		mkTest(
			"junction-basic2",
			func() *Graph {
				e00 := ValueOut(id("v0"), id("e00"))
				e10 := ValueOut(id("v1"), id("e10"))
				ej0 := JunctionOut([]HalfEdge{e00, e10}, id("ej0"))
				e01 := ValueOut(id("v0"), id("e01"))
				e11 := ValueOut(id("v1"), id("e11"))
				ej1 := JunctionOut([]HalfEdge{e01, e11}, id("ej1"))
				ej2 := JunctionOut([]HalfEdge{ej0, ej1}, id("ej2"))
				return Null.AddValueIn(ej2, id("v2"))
			},
			value("v0"), value("v1"),
			value("v2", junction("ej2",
				junction("ej0",
					edge("e00", value("v0")),
					edge("e10", value("v1")),
				),
				junction("ej1",
					edge("e01", value("v0")),
					edge("e11", value("v1")),
				),
			)),
		),

		mkTest(
			"junction-circular",
			func() *Graph {
				e0 := ValueOut(id("v0"), id("e0"))
				e1 := ValueOut(id("v1"), id("e1"))
				ej0 := JunctionOut([]HalfEdge{e0, e1}, id("ej0"))
				g0 := Null.AddValueIn(ej0, id("v2"))
				e20 := ValueOut(id("v2"), id("e20"))
				g1 := g0.AddValueIn(e20, id("v0"))
				e21 := ValueOut(id("v2"), id("e21"))
				return g1.AddValueIn(e21, id("v1"))
			},
			value("v0", edge("e20", value("v2", junction("ej0",
				edge("e0", value("v0")),
				edge("e1", value("v1", edge("e21", value("v2")))),
			)))),
			value("v1", edge("e21", value("v2", junction("ej0",
				edge("e0", value("v0", edge("e20", value("v2")))),
				edge("e1", value("v1")),
			)))),
			value("v2", junction("ej0",
				edge("e0", value("v0", edge("e20", value("v2")))),
				edge("e1", value("v1", edge("e21", value("v2")))),
			)),
		),
	}

	for i := range tests {
		out := tests[i].out()
		for j, exp := range tests[i].exp {
			msgAndArgs := []interface{}{
				"tests[%d].name:%q exp[%d].val:%q",
				i, tests[i].name, j, exp.Value.(idAny).i,
			}
			v := out.Value(exp.Value)
			if !assert.NotNil(t, v, msgAndArgs...) {
				continue
			}
			assertVertexEqual(t, exp, v, msgAndArgs...)
		}

		// sanity check that graphs are equal to themselves
		assert.True(t, Equal(out, out))
	}
}

func TestGraphImmutability(t *T) {
	e0 := ValueOut(id("v0"), id("e0"))
	g0 := Null.AddValueIn(e0, id("v1"))
	assert.Nil(t, Null.Value(id("v0")))
	assert.Nil(t, Null.Value(id("v1")))
	assert.NotNil(t, g0.Value(id("v0")))
	assert.NotNil(t, g0.Value(id("v1")))

	// half-edges should be re-usable
	e1 := ValueOut(id("v2"), id("e1"))
	g1a := g0.AddValueIn(e1, id("v3a"))
	g1b := g0.AddValueIn(e1, id("v3b"))
	assertVertexEqual(t, value("v3a", edge("e1", value("v2"))), g1a.Value(id("v3a")))
	assert.Nil(t, g1a.Value(id("v3b")))
	assertVertexEqual(t, value("v3b", edge("e1", value("v2"))), g1b.Value(id("v3b")))
	assert.Nil(t, g1b.Value(id("v3a")))

	// ... even re-usable twice in succession
	g2 := g0.AddValueIn(e1, id("v3")).AddValueIn(e1, id("v4"))
	assert.Nil(t, g2.Value(id("v3b")))
	assert.Nil(t, g2.Value(id("v3a")))
	assertVertexEqual(t, value("v3", edge("e1", value("v2"))), g2.Value(id("v3")))
	assertVertexEqual(t, value("v4", edge("e1", value("v2"))), g2.Value(id("v4")))
}

func TestGraphDelValueIn(t *T) {
	{ // removing from null
		g := Null.DelValueIn(ValueOut(id("v0"), id("e0")), id("v1"))
		assert.True(t, Equal(Null, g))
	}

	{ // removing edge from vertex which doesn't have that edge
		g0 := Null.AddValueIn(ValueOut(id("v0"), id("e0")), id("v1"))
		g1 := g0.DelValueIn(ValueOut(id("v0"), id("e1")), id("v1"))
		assert.True(t, Equal(g0, g1))
	}

	{ // removing only edge
		he := ValueOut(id("v0"), id("e0"))
		g0 := Null.AddValueIn(he, id("v1"))
		g1 := g0.DelValueIn(he, id("v1"))
		assert.True(t, Equal(Null, g1))
	}

	{ // removing only edge (junction)
		he := JunctionOut([]HalfEdge{
			ValueOut(id("v0"), id("e0")),
			ValueOut(id("v1"), id("e1")),
		}, id("ej0"))
		g0 := Null.AddValueIn(he, id("v2"))
		g1 := g0.DelValueIn(he, id("v2"))
		assert.True(t, Equal(Null, g1))
	}

	{ // removing one of two edges
		he := ValueOut(id("v1"), id("e0"))
		g0 := Null.AddValueIn(ValueOut(id("v0"), id("e0")), id("v2"))
		g1 := g0.AddValueIn(he, id("v2"))
		g2 := g1.DelValueIn(he, id("v2"))
		assert.True(t, Equal(g0, g2))
		assert.NotNil(t, g2.Value(id("v0")))
		assert.Nil(t, g2.Value(id("v1")))
		assert.NotNil(t, g2.Value(id("v2")))
	}

	{ // removing one of two edges (junction)
		e0 := ValueOut(id("v0"), id("e0"))
		e1 := ValueOut(id("v1"), id("e1"))
		e2 := ValueOut(id("v2"), id("e2"))
		heA := JunctionOut([]HalfEdge{e0, e1}, id("heA"))
		heB := JunctionOut([]HalfEdge{e1, e2}, id("heB"))
		g0a := Null.AddValueIn(heA, id("v3"))
		g0b := Null.AddValueIn(heB, id("v3"))
		g1 := g0a.Union(g0b).DelValueIn(heA, id("v3"))
		assert.True(t, Equal(g1, g0b))
		assert.Nil(t, g1.Value(id("v0")))
		assert.NotNil(t, g1.Value(id("v1")))
		assert.NotNil(t, g1.Value(id("v2")))
		assert.NotNil(t, g1.Value(id("v3")))
	}

	{ // removing one of two edges in circular graph
		e0 := ValueOut(id("v0"), id("e0"))
		e1 := ValueOut(id("v1"), id("e1"))
		g0 := Null.AddValueIn(e0, id("v1")).AddValueIn(e1, id("v0"))
		g1 := g0.DelValueIn(e0, id("v1"))
		assert.True(t, Equal(Null.AddValueIn(e1, id("v0")), g1))
		assert.NotNil(t, g1.Value(id("v0")))
		assert.NotNil(t, g1.Value(id("v1")))
	}

	{ // removing to's only edge, sub-nodes have edge to each other
		ej := JunctionOut([]HalfEdge{
			ValueOut(id("v0"), id("ej0")),
			ValueOut(id("v1"), id("ej0")),
		}, id("ej"))
		g0 := Null.AddValueIn(ej, id("v2"))
		e0 := ValueOut(id("v0"), id("e0"))
		g1 := g0.AddValueIn(e0, id("v1"))
		g2 := g1.DelValueIn(ej, id("v2"))
		assert.True(t, Equal(Null.AddValueIn(e0, id("v1")), g2))
		assert.NotNil(t, g2.Value(id("v0")))
		assert.NotNil(t, g2.Value(id("v1")))
		assert.Nil(t, g2.Value(id("v2")))
	}
}

func TestGraphUnion(t *T) {
	assertUnion := func(g1, g2 *Graph) *Graph {
		ga := g1.Union(g2)
		gb := g2.Union(g1)
		assert.True(t, Equal(ga, gb))
		return ga
	}

	{ // Union with Null
		assert.True(t, Equal(Null, Null.Union(Null)))

		g := Null.AddValueIn(ValueOut(id("v0"), id("e0")), id("v1"))
		assert.True(t, Equal(g, assertUnion(g, Null)))
	}

	{ // Two disparate graphs union'd
		g0 := Null.AddValueIn(ValueOut(id("v0"), id("e0")), id("v1"))
		g1 := Null.AddValueIn(ValueOut(id("v2"), id("e1")), id("v3"))
		g := assertUnion(g0, g1)
		assertVertexEqual(t, value("v0"), g.Value(id("v0")))
		assertVertexEqual(t, value("v1", edge("e0", value("v0"))), g.Value(id("v1")))
		assertVertexEqual(t, value("v2"), g.Value(id("v2")))
		assertVertexEqual(t, value("v3", edge("e1", value("v2"))), g.Value(id("v3")))
	}

	{ // Two disparate graphs with junctions
		ga := Null.AddValueIn(JunctionOut([]HalfEdge{
			ValueOut(id("va0"), id("ea0")),
			ValueOut(id("va1"), id("ea1")),
		}, id("eaj")), id("va2"))
		gb := Null.AddValueIn(JunctionOut([]HalfEdge{
			ValueOut(id("vb0"), id("eb0")),
			ValueOut(id("vb1"), id("eb1")),
		}, id("ebj")), id("vb2"))
		g := assertUnion(ga, gb)
		assertVertexEqual(t, value("va0"), g.Value(id("va0")))
		assertVertexEqual(t, value("va1"), g.Value(id("va1")))
		assertVertexEqual(t,
			value("va2", junction("eaj",
				edge("ea0", value("va0")),
				edge("ea1", value("va1")))),
			g.Value(id("va2")),
		)
		assertVertexEqual(t, value("vb0"), g.Value(id("vb0")))
		assertVertexEqual(t, value("vb1"), g.Value(id("vb1")))
		assertVertexEqual(t,
			value("vb2", junction("ebj",
				edge("eb0", value("vb0")),
				edge("eb1", value("vb1")))),
			g.Value(id("vb2")),
		)
	}

	{ // Two partially overlapping graphs
		g0 := Null.AddValueIn(ValueOut(id("v0"), id("e0")), id("v2"))
		g1 := Null.AddValueIn(ValueOut(id("v1"), id("e1")), id("v2"))
		g := assertUnion(g0, g1)
		assertVertexEqual(t, value("v0"), g.Value(id("v0")))
		assertVertexEqual(t, value("v1"), g.Value(id("v1")))
		assertVertexEqual(t,
			value("v2",
				edge("e0", value("v0")),
				edge("e1", value("v1")),
			),
			g.Value(id("v2")),
		)
	}

	{ // two partially overlapping graphs with junctions
		g0 := Null.AddValueIn(JunctionOut([]HalfEdge{
			ValueOut(id("v0"), id("e0")),
			ValueOut(id("v1"), id("e1")),
		}, id("ej0")), id("v2"))
		g1 := Null.AddValueIn(JunctionOut([]HalfEdge{
			ValueOut(id("v0"), id("e0")),
			ValueOut(id("v1"), id("e1")),
		}, id("ej1")), id("v2"))
		g := assertUnion(g0, g1)
		assertVertexEqual(t, value("v0"), g.Value(id("v0")))
		assertVertexEqual(t, value("v1"), g.Value(id("v1")))
		assertVertexEqual(t,
			value("v2",
				junction("ej0", edge("e0", value("v0")), edge("e1", value("v1"))),
				junction("ej1", edge("e0", value("v0")), edge("e1", value("v1"))),
			),
			g.Value(id("v2")),
		)
	}

	{ // Two equal graphs
		g0 := Null.AddValueIn(ValueOut(id("v0"), id("e0")), id("v1"))
		g := assertUnion(g0, g0)
		assertVertexEqual(t, value("v0"), g.Value(id("v0")))
		assertVertexEqual(t,
			value("v1", edge("e0", value("v0"))),
			g.Value(id("v1")),
		)
	}

	{ // Two equal graphs with junctions
		g0 := Null.AddValueIn(JunctionOut([]HalfEdge{
			ValueOut(id("v0"), id("e0")),
			ValueOut(id("v1"), id("e1")),
		}, id("ej0")), id("v2"))
		g := assertUnion(g0, g0)
		assertVertexEqual(t, value("v0"), g.Value(id("v0")))
		assertVertexEqual(t, value("v1"), g.Value(id("v1")))
		assertVertexEqual(t,
			value("v2",
				junction("ej0", edge("e0", value("v0")), edge("e1", value("v1"))),
			),
			g.Value(id("v2")),
		)
	}
}

func TestGraphEqual(t *T) {
	assertEqual := func(g1, g2 *Graph) {
		assert.True(t, Equal(g1, g2))
		assert.True(t, Equal(g2, g1))
	}

	assertNotEqual := func(g1, g2 *Graph) {
		assert.False(t, Equal(g1, g2))
		assert.False(t, Equal(g2, g1))
	}

	assertEqual(Null, Null) // duh

	{
		// graph is equal to itself, not to null
		e0 := ValueOut(id("v0"), id("e0"))
		g0 := Null.AddValueIn(e0, id("v1"))
		assertNotEqual(g0, Null)
		assertEqual(g0, g0)

		// adding the an existing edge again shouldn't do anything
		assertEqual(g0, g0.AddValueIn(e0, id("v1")))

		// g1a and g1b have the same vertices, but the edges are different
		g1a := g0.AddValueIn(ValueOut(id("v0"), id("e1a")), id("v2"))
		g1b := g0.AddValueIn(ValueOut(id("v0"), id("e1b")), id("v2"))
		assertNotEqual(g1a, g1b)
	}

	{ // equal construction should yield equality, even if out of order
		ga := Null.AddValueIn(ValueOut(id("v0"), id("e0")), id("v1"))
		ga = ga.AddValueIn(ValueOut(id("v1"), id("e1")), id("v2"))
		gb := Null.AddValueIn(ValueOut(id("v1"), id("e1")), id("v2"))
		gb = gb.AddValueIn(ValueOut(id("v0"), id("e0")), id("v1"))
		assertEqual(ga, gb)
	}

	{ // junction basic test
		e0 := ValueOut(id("v0"), id("e0"))
		e1 := ValueOut(id("v1"), id("e1"))
		ga := Null.AddValueIn(JunctionOut([]HalfEdge{e0, e1}, id("ej")), id("v2"))
		gb := Null.AddValueIn(JunctionOut([]HalfEdge{e1, e0}, id("ej")), id("v2"))
		assertEqual(ga, ga)
		assertNotEqual(ga, gb)
	}
}
