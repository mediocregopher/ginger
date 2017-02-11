package lang

import "fmt"

// Match is used to pattern match an arbitrary Term against a pattern. A pattern
// is a 2-tuple of the type (as an atom, e.g. AAtom, AConst) and a matching
// value.
//
// If the value is AUnder the pattern will match all Terms of the type,
// regardless of their value. If the pattern's type and value are both AUnder
// the pattern will match all Terms.
//
// If the pattern's value is a Tuple or a List, each of its values will be used
// as a sub-pattern to match against the corresponding value in the value.
func Match(pat Tuple, t Term) bool {
	if len(pat) != 2 {
		return false
	}
	pt, pv := pat[0], pat[1]

	switch pt {
	case AAtom:
		a, ok := t.(Atom)
		return ok && (Equal(pv, AUnder) || Equal(pv, a))
	case AConst:
		c, ok := t.(Const)
		return ok && (Equal(pv, AUnder) || Equal(pv, c))
	case ATuple:
		tt, ok := t.(Tuple)
		if !ok {
			return false
		} else if Equal(pv, AUnder) {
			return true
		}

		pvt := pv.(Tuple)
		if len(tt) != len(pvt) {
			return false
		}
		for i := range tt {
			pvti, ok := pvt[i].(Tuple)
			if !ok || !Match(pvti, tt[i]) {
				return false
			}
		}
		return true
	case AList:
		panic("TODO")
	case AUnder:
		return true
	default:
		panic(fmt.Sprintf("unknown type %T", pt))
	}
}
