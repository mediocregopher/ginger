package gg

import (
	"fmt"
	"sort"
	"strings"
	. "testing"

	"github.com/stretchr/testify/assert"
)

func edge(val Value, from *Vertex) Edge {
	return Edge{Value: val, From: from}
}

func value(val Value, in ...Edge) *Vertex {
	return &Vertex{
		VertexType: ValueVertex,
		Value:      val,
		In:         in,
	}
}

func junction(val Value, in ...Edge) Edge {
	return Edge{
		From: &Vertex{
			VertexType: JunctionVertex,
			In:         in,
		},
		Value: val,
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

func assertIter(t *T, expVals, expJuncs int, g *Graph, msgAndArgs ...interface{}) {
	seen := map[*Vertex]bool{}
	var gotVals, gotJuncs int
	g.Iter(func(v *Vertex) bool {
		assert.NotContains(t, seen, v, msgAndArgs...)
		seen[v] = true
		if v.VertexType == ValueVertex {
			gotVals++
		} else {
			gotJuncs++
		}
		return true
	})
	assert.Equal(t, expVals, gotVals, msgAndArgs...)
	assert.Equal(t, expJuncs, gotJuncs, msgAndArgs...)
}

type graphTest struct {
	name              string
	out               func() *Graph
	exp               []*Vertex
	numVals, numJuncs int
}

func mkTest(name string, out func() *Graph, numVals, numJuncs int, exp ...*Vertex) graphTest {
	return graphTest{
		name:    name,
		out:     out,
		exp:     exp,
		numVals: numVals, numJuncs: numJuncs,
	}
}

func TestGraph(t *T) {
	var (
		v0  = NewValue("v0")
		v1  = NewValue("v1")
		v2  = NewValue("v2")
		v3  = NewValue("v3")
		e0  = NewValue("e0")
		e00 = NewValue("e00")
		e01 = NewValue("e01")
		e1  = NewValue("e1")
		e10 = NewValue("e10")
		e11 = NewValue("e11")
		e2  = NewValue("e2")
		e20 = NewValue("e20")
		e21 = NewValue("e21")
		ej0 = NewValue("ej0")
		ej1 = NewValue("ej1")
		ej2 = NewValue("ej2")
	)
	tests := []graphTest{
		mkTest(
			"values-basic",
			func() *Graph {
				return Null.AddValueIn(ValueOut(v0, e0), v1)
			},
			2, 0,
			value(v0),
			value(v1, edge(e0, value(v0))),
		),

		mkTest(
			"values-2edges",
			func() *Graph {
				g0 := Null.AddValueIn(ValueOut(v0, e0), v2)
				return g0.AddValueIn(ValueOut(v1, e1), v2)
			},
			3, 0,
			value(v0),
			value(v1),
			value(v2,
				edge(e0, value(v0)),
				edge(e1, value(v1)),
			),
		),

		mkTest(
			"values-separate",
			func() *Graph {
				g0 := Null.AddValueIn(ValueOut(v0, e0), v1)
				return g0.AddValueIn(ValueOut(v2, e2), v3)
			},
			4, 0,
			value(v0),
			value(v1, edge(e0, value(v0))),
			value(v2),
			value(v3, edge(e2, value(v2))),
		),

		mkTest(
			"values-circular",
			func() *Graph {
				return Null.AddValueIn(ValueOut(v0, e0), v0)
			},
			1, 0,
			value(v0, edge(e0, value(v0))),
		),

		mkTest(
			"values-circular2",
			func() *Graph {
				g0 := Null.AddValueIn(ValueOut(v0, e0), v1)
				return g0.AddValueIn(ValueOut(v1, e1), v0)
			},
			2, 0,
			value(v0, edge(e1, value(v1, edge(e0, value(v0))))),
			value(v1, edge(e0, value(v0, edge(e1, value(v1))))),
		),

		mkTest(
			"values-circular3",
			func() *Graph {
				g0 := Null.AddValueIn(ValueOut(v0, e0), v1)
				g1 := g0.AddValueIn(ValueOut(v1, e1), v2)
				return g1.AddValueIn(ValueOut(v2, e2), v1)
			},
			3, 0,
			value(v0),
			value(v1,
				edge(e0, value(v0)),
				edge(e2, value(v2, edge(e1, value(v1)))),
			),
			value(v2, edge(e1, value(v1,
				edge(e0, value(v0)),
				edge(e2, value(v2)),
			))),
		),

		mkTest(
			"junction-basic",
			func() *Graph {
				e0 := ValueOut(v0, e0)
				e1 := ValueOut(v1, e1)
				ej0 := JunctionOut([]OpenEdge{e0, e1}, ej0)
				return Null.AddValueIn(ej0, v2)
			},
			3, 1,
			value(v0), value(v1),
			value(v2, junction(ej0,
				edge(e0, value(v0)),
				edge(e1, value(v1)),
			)),
		),

		mkTest(
			"junction-basic2",
			func() *Graph {
				e00 := ValueOut(v0, e00)
				e10 := ValueOut(v1, e10)
				ej0 := JunctionOut([]OpenEdge{e00, e10}, ej0)
				e01 := ValueOut(v0, e01)
				e11 := ValueOut(v1, e11)
				ej1 := JunctionOut([]OpenEdge{e01, e11}, ej1)
				ej2 := JunctionOut([]OpenEdge{ej0, ej1}, ej2)
				return Null.AddValueIn(ej2, v2)
			},
			3, 3,
			value(v0), value(v1),
			value(v2, junction(ej2,
				junction(ej0,
					edge(e00, value(v0)),
					edge(e10, value(v1)),
				),
				junction(ej1,
					edge(e01, value(v0)),
					edge(e11, value(v1)),
				),
			)),
		),

		mkTest(
			"junction-circular",
			func() *Graph {
				e0 := ValueOut(v0, e0)
				e1 := ValueOut(v1, e1)
				ej0 := JunctionOut([]OpenEdge{e0, e1}, ej0)
				g0 := Null.AddValueIn(ej0, v2)
				e20 := ValueOut(v2, e20)
				g1 := g0.AddValueIn(e20, v0)
				e21 := ValueOut(v2, e21)
				return g1.AddValueIn(e21, v1)
			},
			3, 1,
			value(v0, edge(e20, value(v2, junction(ej0,
				edge(e0, value(v0)),
				edge(e1, value(v1, edge(e21, value(v2)))),
			)))),
			value(v1, edge(e21, value(v2, junction(ej0,
				edge(e0, value(v0, edge(e20, value(v2)))),
				edge(e1, value(v1)),
			)))),
			value(v2, junction(ej0,
				edge(e0, value(v0, edge(e20, value(v2)))),
				edge(e1, value(v1, edge(e21, value(v2)))),
			)),
		),
	}

	for i := range tests {
		t.Logf("test[%d]:%q", i, tests[i].name)
		out := tests[i].out()
		for j, exp := range tests[i].exp {
			msgAndArgs := []interface{}{
				"tests[%d].name:%q exp[%d].val:%q",
				i, tests[i].name, j, exp.Value.V.(string),
			}
			v := out.ValueVertex(exp.Value)
			if !assert.NotNil(t, v, msgAndArgs...) {
				continue
			}
			assertVertexEqual(t, exp, v, msgAndArgs...)
		}

		msgAndArgs := []interface{}{
			"tests[%d].name:%q",
			i, tests[i].name,
		}

		// sanity check that graphs are equal to themselves
		assert.True(t, Equal(out, out), msgAndArgs...)

		// test the Iter method in here too
		assertIter(t, tests[i].numVals, tests[i].numJuncs, out, msgAndArgs...)
	}
}

func TestGraphImmutability(t *T) {
	v0 := NewValue("v0")
	v1 := NewValue("v1")
	e0 := NewValue("e0")
	oe0 := ValueOut(v0, e0)
	g0 := Null.AddValueIn(oe0, v1)
	assert.Nil(t, Null.ValueVertex(v0))
	assert.Nil(t, Null.ValueVertex(v1))
	assert.NotNil(t, g0.ValueVertex(v0))
	assert.NotNil(t, g0.ValueVertex(v1))

	// half-edges should be re-usable
	v2 := NewValue("v2")
	v3a, v3b := NewValue("v3a"), NewValue("v3b")
	e1 := NewValue("e1")
	oe1 := ValueOut(v2, e1)
	g1a := g0.AddValueIn(oe1, v3a)
	g1b := g0.AddValueIn(oe1, v3b)
	assertVertexEqual(t, value(v3a, edge(e1, value(v2))), g1a.ValueVertex(v3a))
	assert.Nil(t, g1a.ValueVertex(v3b))
	assertVertexEqual(t, value(v3b, edge(e1, value(v2))), g1b.ValueVertex(v3b))
	assert.Nil(t, g1b.ValueVertex(v3a))

	// ... even re-usable twice in succession
	v3 := NewValue("v3")
	v4 := NewValue("v4")
	g2 := g0.AddValueIn(oe1, v3).AddValueIn(oe1, v4)
	assert.Nil(t, g2.ValueVertex(v3b))
	assert.Nil(t, g2.ValueVertex(v3a))
	assertVertexEqual(t, value(v3, edge(e1, value(v2))), g2.ValueVertex(v3))
	assertVertexEqual(t, value(v4, edge(e1, value(v2))), g2.ValueVertex(v4))
}

func TestGraphDelValueIn(t *T) {
	v0 := NewValue("v0")
	v1 := NewValue("v1")
	e0 := NewValue("e0")
	{ // removing from null
		g := Null.DelValueIn(ValueOut(v0, e0), v1)
		assert.True(t, Equal(Null, g))
	}

	e1 := NewValue("e1")
	{ // removing edge from vertex which doesn't have that edge
		g0 := Null.AddValueIn(ValueOut(v0, e0), v1)
		g1 := g0.DelValueIn(ValueOut(v0, e1), v1)
		assert.True(t, Equal(g0, g1))
	}

	{ // removing only edge
		oe := ValueOut(v0, e0)
		g0 := Null.AddValueIn(oe, v1)
		g1 := g0.DelValueIn(oe, v1)
		assert.True(t, Equal(Null, g1))
	}

	ej0 := NewValue("ej0")
	v2 := NewValue("v2")
	{ // removing only edge (junction)
		oe := JunctionOut([]OpenEdge{
			ValueOut(v0, e0),
			ValueOut(v1, e1),
		}, ej0)
		g0 := Null.AddValueIn(oe, v2)
		g1 := g0.DelValueIn(oe, v2)
		assert.True(t, Equal(Null, g1))
	}

	{ // removing one of two edges
		oe := ValueOut(v1, e0)
		g0 := Null.AddValueIn(ValueOut(v0, e0), v2)
		g1 := g0.AddValueIn(oe, v2)
		g2 := g1.DelValueIn(oe, v2)
		assert.True(t, Equal(g0, g2))
		assert.NotNil(t, g2.ValueVertex(v0))
		assert.Nil(t, g2.ValueVertex(v1))
		assert.NotNil(t, g2.ValueVertex(v2))
	}

	e2 := NewValue("e2")
	eja, ejb := NewValue("eja"), NewValue("ejb")
	v3 := NewValue("v3")
	{ // removing one of two edges (junction)
		e0 := ValueOut(v0, e0)
		e1 := ValueOut(v1, e1)
		e2 := ValueOut(v2, e2)
		oeA := JunctionOut([]OpenEdge{e0, e1}, eja)
		oeB := JunctionOut([]OpenEdge{e1, e2}, ejb)
		g0a := Null.AddValueIn(oeA, v3)
		g0b := Null.AddValueIn(oeB, v3)
		g1 := g0a.Union(g0b).DelValueIn(oeA, v3)
		assert.True(t, Equal(g1, g0b))
		assert.Nil(t, g1.ValueVertex(v0))
		assert.NotNil(t, g1.ValueVertex(v1))
		assert.NotNil(t, g1.ValueVertex(v2))
		assert.NotNil(t, g1.ValueVertex(v3))
	}

	{ // removing one of two edges in circular graph
		e0 := ValueOut(v0, e0)
		e1 := ValueOut(v1, e1)
		g0 := Null.AddValueIn(e0, v1).AddValueIn(e1, v0)
		g1 := g0.DelValueIn(e0, v1)
		assert.True(t, Equal(Null.AddValueIn(e1, v0), g1))
		assert.NotNil(t, g1.ValueVertex(v0))
		assert.NotNil(t, g1.ValueVertex(v1))
	}

	ej := NewValue("ej")
	{ // removing to's only edge, sub-nodes have edge to each other
		oej := JunctionOut([]OpenEdge{
			ValueOut(v0, ej0),
			ValueOut(v1, ej0),
		}, ej)
		g0 := Null.AddValueIn(oej, v2)
		e0 := ValueOut(v0, e0)
		g1 := g0.AddValueIn(e0, v1)
		g2 := g1.DelValueIn(oej, v2)
		assert.True(t, Equal(Null.AddValueIn(e0, v1), g2))
		assert.NotNil(t, g2.ValueVertex(v0))
		assert.NotNil(t, g2.ValueVertex(v1))
		assert.Nil(t, g2.ValueVertex(v2))
	}
}

// deterministically hashes a Graph
func graphStr(g *Graph) string {
	var vStr func(vertex) string
	var oeStr func(OpenEdge) string
	vStr = func(v vertex) string {
		if v.VertexType == ValueVertex {
			return fmt.Sprintf("v:%q\n", v.val.V.(string))
		}
		s := fmt.Sprintf("j:%d\n", len(v.in))
		ssOE := make([]string, len(v.in))
		for i := range v.in {
			ssOE[i] = oeStr(v.in[i])
		}
		sort.Strings(ssOE)
		return s + strings.Join(ssOE, "")
	}
	oeStr = func(oe OpenEdge) string {
		s := fmt.Sprintf("oe:%q\n", oe.val.V.(string))
		return s + vStr(oe.fromV)
	}
	sVV := make([]string, 0, len(g.vM))
	for _, v := range g.vM {
		sVV = append(sVV, vStr(v))
	}
	sort.Strings(sVV)
	return strings.Join(sVV, "")
}

func assertEqualSets(t *T, exp, got []*Graph) bool {
	if !assert.Equal(t, len(exp), len(got)) {
		return false
	}

	m := map[*Graph]string{}
	for _, g := range exp {
		m[g] = graphStr(g)
	}
	for _, g := range got {
		m[g] = graphStr(g)
	}

	sort.Slice(exp, func(i, j int) bool {
		return m[exp[i]] < m[exp[j]]
	})
	sort.Slice(got, func(i, j int) bool {
		return m[got[i]] < m[got[j]]
	})

	b := true
	for i := range exp {
		b = b || assert.True(t, Equal(exp[i], got[i]), "i:%d exp:%q got:%q", i, m[exp[i]], m[got[i]])
	}
	return b
}

func TestGraphUnion(t *T) {
	assertUnion := func(g1, g2 *Graph) *Graph {
		ga := g1.Union(g2)
		gb := g2.Union(g1)
		assert.True(t, Equal(ga, gb))
		return ga
	}

	assertDisjoin := func(g *Graph, exp ...*Graph) {
		ggDisj := g.Disjoin()
		assertEqualSets(t, exp, ggDisj)
	}

	v0 := NewValue("v0")
	v1 := NewValue("v1")
	e0 := NewValue("e0")
	{ // Union with Null
		assert.True(t, Equal(Null, Null.Union(Null)))

		g := Null.AddValueIn(ValueOut(v0, e0), v1)
		assert.True(t, Equal(g, assertUnion(g, Null)))

		assertDisjoin(g, g)
	}

	v2 := NewValue("v2")
	v3 := NewValue("v3")
	e1 := NewValue("e1")
	{ // Two disparate graphs union'd
		g0 := Null.AddValueIn(ValueOut(v0, e0), v1)
		g1 := Null.AddValueIn(ValueOut(v2, e1), v3)
		g := assertUnion(g0, g1)
		assertVertexEqual(t, value(v0), g.ValueVertex(v0))
		assertVertexEqual(t, value(v1, edge(e0, value(v0))), g.ValueVertex(v1))
		assertVertexEqual(t, value(v2), g.ValueVertex(v2))
		assertVertexEqual(t, value(v3, edge(e1, value(v2))), g.ValueVertex(v3))

		assertDisjoin(g, g0, g1)
	}

	va0, vb0 := NewValue("va0"), NewValue("vb0")
	va1, vb1 := NewValue("va1"), NewValue("vb1")
	va2, vb2 := NewValue("va2"), NewValue("vb2")
	ea0, eb0 := NewValue("ea0"), NewValue("eb0")
	ea1, eb1 := NewValue("ea1"), NewValue("eb1")
	eaj, ebj := NewValue("eaj"), NewValue("ebj")
	{ // Two disparate graphs with junctions
		ga := Null.AddValueIn(JunctionOut([]OpenEdge{
			ValueOut(va0, ea0),
			ValueOut(va1, ea1),
		}, eaj), va2)
		gb := Null.AddValueIn(JunctionOut([]OpenEdge{
			ValueOut(vb0, eb0),
			ValueOut(vb1, eb1),
		}, ebj), vb2)
		g := assertUnion(ga, gb)
		assertVertexEqual(t, value(va0), g.ValueVertex(va0))
		assertVertexEqual(t, value(va1), g.ValueVertex(va1))
		assertVertexEqual(t,
			value(va2, junction(eaj,
				edge(ea0, value(va0)),
				edge(ea1, value(va1)))),
			g.ValueVertex(va2),
		)
		assertVertexEqual(t, value(vb0), g.ValueVertex(vb0))
		assertVertexEqual(t, value(vb1), g.ValueVertex(vb1))
		assertVertexEqual(t,
			value(vb2, junction(ebj,
				edge(eb0, value(vb0)),
				edge(eb1, value(vb1)))),
			g.ValueVertex(vb2),
		)

		assertDisjoin(g, ga, gb)
	}

	{ // Two partially overlapping graphs
		g0 := Null.AddValueIn(ValueOut(v0, e0), v2)
		g1 := Null.AddValueIn(ValueOut(v1, e1), v2)
		g := assertUnion(g0, g1)
		assertVertexEqual(t, value(v0), g.ValueVertex(v0))
		assertVertexEqual(t, value(v1), g.ValueVertex(v1))
		assertVertexEqual(t,
			value(v2,
				edge(e0, value(v0)),
				edge(e1, value(v1)),
			),
			g.ValueVertex(v2),
		)

		assertDisjoin(g, g)
	}

	ej0 := NewValue("ej0")
	ej1 := NewValue("ej1")
	{ // two partially overlapping graphs with junctions
		g0 := Null.AddValueIn(JunctionOut([]OpenEdge{
			ValueOut(v0, e0),
			ValueOut(v1, e1),
		}, ej0), v2)
		g1 := Null.AddValueIn(JunctionOut([]OpenEdge{
			ValueOut(v0, e0),
			ValueOut(v1, e1),
		}, ej1), v2)
		g := assertUnion(g0, g1)
		assertVertexEqual(t, value(v0), g.ValueVertex(v0))
		assertVertexEqual(t, value(v1), g.ValueVertex(v1))
		assertVertexEqual(t,
			value(v2,
				junction(ej0, edge(e0, value(v0)), edge(e1, value(v1))),
				junction(ej1, edge(e0, value(v0)), edge(e1, value(v1))),
			),
			g.ValueVertex(v2),
		)

		assertDisjoin(g, g)
	}

	{ // Two equal graphs
		g0 := Null.AddValueIn(ValueOut(v0, e0), v1)
		g := assertUnion(g0, g0)
		assertVertexEqual(t, value(v0), g.ValueVertex(v0))
		assertVertexEqual(t,
			value(v1, edge(e0, value(v0))),
			g.ValueVertex(v1),
		)
	}

	{ // Two equal graphs with junctions
		g0 := Null.AddValueIn(JunctionOut([]OpenEdge{
			ValueOut(v0, e0),
			ValueOut(v1, e1),
		}, ej0), v2)
		g := assertUnion(g0, g0)
		assertVertexEqual(t, value(v0), g.ValueVertex(v0))
		assertVertexEqual(t, value(v1), g.ValueVertex(v1))
		assertVertexEqual(t,
			value(v2,
				junction(ej0, edge(e0, value(v0)), edge(e1, value(v1))),
			),
			g.ValueVertex(v2),
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

	v0 := NewValue("v0")
	v1 := NewValue("v1")
	v2 := NewValue("v2")
	e0 := NewValue("e0")
	e1 := NewValue("e1")
	e1a, e1b := NewValue("e1a"), NewValue("e1b")
	{
		// graph is equal to itself, not to null
		e0 := ValueOut(v0, e0)
		g0 := Null.AddValueIn(e0, v1)
		assertNotEqual(g0, Null)
		assertEqual(g0, g0)

		// adding the an existing edge again shouldn't do anything
		assertEqual(g0, g0.AddValueIn(e0, v1))

		// g1a and g1b have the same vertices, but the edges are different
		g1a := g0.AddValueIn(ValueOut(v0, e1a), v2)
		g1b := g0.AddValueIn(ValueOut(v0, e1b), v2)
		assertNotEqual(g1a, g1b)
	}

	{ // equal construction should yield equality, even if out of order
		ga := Null.AddValueIn(ValueOut(v0, e0), v1)
		ga = ga.AddValueIn(ValueOut(v1, e1), v2)
		gb := Null.AddValueIn(ValueOut(v1, e1), v2)
		gb = gb.AddValueIn(ValueOut(v0, e0), v1)
		assertEqual(ga, gb)
	}

	ej := NewValue("ej")
	{ // junction basic test
		e0 := ValueOut(v0, e0)
		e1 := ValueOut(v1, e1)
		ga := Null.AddValueIn(JunctionOut([]OpenEdge{e0, e1}, ej), v2)
		gb := Null.AddValueIn(JunctionOut([]OpenEdge{e1, e0}, ej), v2)
		assertEqual(ga, ga)
		assertNotEqual(ga, gb)
	}
}
