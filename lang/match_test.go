package lang

import (
	. "testing"

	"github.com/stretchr/testify/assert"
)

func TestMatch(t *T) {
	pat := func(typ, val Term) Tuple {
		return Tuple{typ, val}
	}

	tests := []struct {
		pattern Tuple
		t       Term
		exp     bool
	}{
		{pat(AAtom, Atom("foo")), Atom("foo"), true},
		{pat(AAtom, Atom("foo")), Atom("bar"), false},
		{pat(AAtom, Atom("foo")), Const("foo"), false},
		{pat(AAtom, Atom("foo")), Tuple{Atom("a"), Atom("b")}, false},
		{pat(AAtom, Atom("_")), Atom("bar"), true},
		{pat(AAtom, Atom("_")), Const("bar"), false},

		{pat(AConst, Const("foo")), Const("foo"), true},
		{pat(AConst, Const("foo")), Atom("foo"), false},
		{pat(AConst, Const("foo")), Const("bar"), false},
		{pat(AConst, Atom("_")), Const("bar"), true},
		{pat(AConst, Atom("_")), Atom("foo"), false},

		{
			pat(ATuple, Tuple{
				pat(AAtom, Atom("foo")),
				pat(AAtom, Atom("bar")),
			}),
			Tuple{Atom("foo"), Atom("bar")},
			true,
		},
		{
			pat(ATuple, Tuple{
				pat(AAtom, Atom("_")),
				pat(AAtom, Atom("bar")),
			}),
			Tuple{Atom("foo"), Atom("bar")},
			true,
		},
		{
			pat(ATuple, Tuple{
				pat(AAtom, Atom("_")),
				pat(AAtom, Atom("_")),
				pat(AAtom, Atom("_")),
			}),
			Tuple{Atom("foo"), Atom("bar")},
			false,
		},

		{pat(AUnder, AUnder), Atom("foo"), true},
		{pat(AUnder, AUnder), Const("foo"), true},
		{pat(AUnder, AUnder), Tuple{Atom("a"), Atom("b")}, true},
	}

	for _, testCase := range tests {
		assert.Equal(t, testCase.exp, Match(testCase.pattern, testCase.t), "%#v", testCase)
	}
}
