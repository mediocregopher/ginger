package constraint

import (
	. "testing"

	"github.com/stretchr/testify/assert"
)

func TestEngineAddConstraint(t *T) {
	{
		e := NewEngine()
		assert.True(t, e.AddConstraint(Constraint{Elem: "0", LT: "1"}))
		assert.True(t, e.AddConstraint(Constraint{Elem: "1", LT: "2"}))
		assert.True(t, e.AddConstraint(Constraint{Elem: "-1", LT: "0"}))
		assert.False(t, e.AddConstraint(Constraint{Elem: "1", LT: "0"}))
		assert.False(t, e.AddConstraint(Constraint{Elem: "2", LT: "0"}))
		assert.False(t, e.AddConstraint(Constraint{Elem: "2", LT: "-1"}))
	}

	{
		e := NewEngine()
		assert.True(t, e.AddConstraint(Constraint{Elem: "0", LT: "1"}))
		assert.True(t, e.AddConstraint(Constraint{Elem: "0", LT: "2"}))
		assert.True(t, e.AddConstraint(Constraint{Elem: "1", LT: "2"}))
		assert.True(t, e.AddConstraint(Constraint{Elem: "2", LT: "3"}))
	}
}

func TestEngineSolve(t *T) {
	assertSolve := func(exp map[string]int, cc ...Constraint) {
		e := NewEngine()
		for _, c := range cc {
			assert.True(t, e.AddConstraint(c), "c:%#v", c)
		}
		assert.Equal(t, exp, e.Solve())
	}

	// basic
	assertSolve(
		map[string]int{"a": 0, "b": 1, "c": 2},
		Constraint{Elem: "a", LT: "b"},
		Constraint{Elem: "b", LT: "c"},
	)

	// "triangle" graph
	assertSolve(
		map[string]int{"a": 0, "b": 1, "c": 2},
		Constraint{Elem: "a", LT: "b"},
		Constraint{Elem: "a", LT: "c"},
		Constraint{Elem: "b", LT: "c"},
	)

	// "hexagon" graph
	assertSolve(
		map[string]int{"a": 0, "b": 1, "c": 1, "d": 2, "e": 2, "f": 3},
		Constraint{Elem: "a", LT: "b"},
		Constraint{Elem: "a", LT: "c"},
		Constraint{Elem: "b", LT: "d"},
		Constraint{Elem: "c", LT: "e"},
		Constraint{Elem: "d", LT: "f"},
		Constraint{Elem: "e", LT: "f"},
	)

	// "hexagon" with centerpoint graph
	assertSolve(
		map[string]int{"a": 0, "b": 1, "c": 1, "center": 2, "d": 3, "e": 3, "f": 4},
		Constraint{Elem: "a", LT: "b"},
		Constraint{Elem: "a", LT: "c"},
		Constraint{Elem: "b", LT: "d"},
		Constraint{Elem: "c", LT: "e"},
		Constraint{Elem: "d", LT: "f"},
		Constraint{Elem: "e", LT: "f"},

		Constraint{Elem: "c", LT: "center"},
		Constraint{Elem: "b", LT: "center"},
		Constraint{Elem: "center", LT: "e"},
		Constraint{Elem: "center", LT: "d"},
	)

	// multi-root, using two triangles which end up connecting
	assertSolve(
		map[string]int{"a": 0, "b": 1, "c": 2, "d": 0, "e": 1, "f": 2, "g": 3},
		Constraint{Elem: "a", LT: "b"},
		Constraint{Elem: "a", LT: "c"},
		Constraint{Elem: "b", LT: "c"},

		Constraint{Elem: "d", LT: "e"},
		Constraint{Elem: "d", LT: "f"},
		Constraint{Elem: "e", LT: "f"},

		Constraint{Elem: "f", LT: "g"},
	)

}
