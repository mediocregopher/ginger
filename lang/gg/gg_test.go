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
				return Null.ValueIn(ValueOut(id("v0"), id("e0")), id("v1"))
			},
			value("v0"),
			value("v1", edge("e0", value("v0"))),
		),

		mkTest(
			"values-2edges",
			func() *Graph {
				g0 := Null.ValueIn(ValueOut(id("v0"), id("e0")), id("v2"))
				return g0.ValueIn(ValueOut(id("v1"), id("e1")), id("v2"))
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
				g0 := Null.ValueIn(ValueOut(id("v0"), id("e0")), id("v1"))
				return g0.ValueIn(ValueOut(id("v2"), id("e2")), id("v3"))
			},
			value("v0"),
			value("v1", edge("e0", value("v0"))),
			value("v2"),
			value("v3", edge("e2", value("v2"))),
		),

		mkTest(
			"values-circular",
			func() *Graph {
				return Null.ValueIn(ValueOut(id("v0"), id("e")), id("v0"))
			},
			value("v0", edge("e", value("v0"))),
		),

		mkTest(
			"values-circular2",
			func() *Graph {
				g0 := Null.ValueIn(ValueOut(id("v0"), id("e0")), id("v1"))
				return g0.ValueIn(ValueOut(id("v1"), id("e1")), id("v0"))
			},
			value("v0", edge("e1", value("v1", edge("e0", value("v0"))))),
			value("v1", edge("e0", value("v0", edge("e1", value("v1"))))),
		),

		mkTest(
			"values-circular3",
			func() *Graph {
				g0 := Null.ValueIn(ValueOut(id("v0"), id("e0")), id("v1"))
				g1 := g0.ValueIn(ValueOut(id("v1"), id("e1")), id("v2"))
				return g1.ValueIn(ValueOut(id("v2"), id("e2")), id("v1"))
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
				return Null.ValueIn(ej0, id("v2"))
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
				return Null.ValueIn(ej2, id("v2"))
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
				g0 := Null.ValueIn(ej0, id("v2"))
				e20 := ValueOut(id("v2"), id("e20"))
				g1 := g0.ValueIn(e20, id("v0"))
				e21 := ValueOut(id("v2"), id("e21"))
				return g1.ValueIn(e21, id("v1"))
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
	g0 := Null.ValueIn(e0, id("v1"))
	assert.Nil(t, Null.Value(id("v0")))
	assert.Nil(t, Null.Value(id("v1")))
	assert.NotNil(t, g0.Value(id("v0")))
	assert.NotNil(t, g0.Value(id("v1")))

	// half-edges should be re-usable
	e1 := ValueOut(id("v2"), id("e1"))
	g1a := g0.ValueIn(e1, id("v3a"))
	g1b := g0.ValueIn(e1, id("v3b"))
	assertVertexEqual(t, value("v3a", edge("e1", value("v2"))), g1a.Value(id("v3a")))
	assert.Nil(t, g1a.Value(id("v3b")))
	assertVertexEqual(t, value("v3b", edge("e1", value("v2"))), g1b.Value(id("v3b")))
	assert.Nil(t, g1b.Value(id("v3a")))

	// ... even re-usable twice in succession
	g2 := g0.ValueIn(e1, id("v3")).ValueIn(e1, id("v4"))
	assert.Nil(t, g2.Value(id("v3b")))
	assert.Nil(t, g2.Value(id("v3a")))
	assertVertexEqual(t, value("v3", edge("e1", value("v2"))), g2.Value(id("v3")))
	assertVertexEqual(t, value("v4", edge("e1", value("v2"))), g2.Value(id("v4")))
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
		g0 := Null.ValueIn(e0, id("v1"))
		assertNotEqual(g0, Null)
		assertEqual(g0, g0)

		// adding the an existing edge again shouldn't do anything
		assertEqual(g0, g0.ValueIn(e0, id("v1")))

		// g1a and g1b have the same vertices, but the edges are different
		g1a := g0.ValueIn(ValueOut(id("v0"), id("e1a")), id("v2"))
		g1b := g0.ValueIn(ValueOut(id("v0"), id("e1b")), id("v2"))
		assertNotEqual(g1a, g1b)
	}

	{ // equal construction should yield equality, even if out of order
		ga := Null.ValueIn(ValueOut(id("v0"), id("e0")), id("v1"))
		ga = ga.ValueIn(ValueOut(id("v1"), id("e1")), id("v2"))
		gb := Null.ValueIn(ValueOut(id("v1"), id("e1")), id("v2"))
		gb = gb.ValueIn(ValueOut(id("v0"), id("e0")), id("v1"))
		assertEqual(ga, gb)
	}

	{ // junction basic test
		e0 := ValueOut(id("v0"), id("e0"))
		e1 := ValueOut(id("v1"), id("e1"))
		ga := Null.ValueIn(JunctionOut([]HalfEdge{e0, e1}, id("ej")), id("v2"))
		gb := Null.ValueIn(JunctionOut([]HalfEdge{e1, e0}, id("ej")), id("v2"))
		assertEqual(ga, ga)
		assertNotEqual(ga, gb)
	}
}
