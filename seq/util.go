package seq

import (
	"testing"

	"github.com/mediocregopher/ginger/types"
)

// Returns whether or not two types.Elem slices contain the same elements
func intSlicesEq(a, b []types.Elem) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func elemSliceV(a ...interface{}) []types.Elem {
	ret := make([]types.Elem, 0, len(a))
	for i := range a {
		if e, ok := a[i].(types.Elem); ok {
			ret = append(ret, e)
		} else {
			ret = append(ret, types.GoType{a[i]})
		}
	}
	return ret
}

// Asserts that the given Seq is empty (contains no elements)
func assertEmpty(s Seq, t *testing.T) {
	if Size(s) != 0 {
		t.Fatalf("Seq isn't empty: %v", ToSlice(s))
	}
}

// Asserts that the given Seq has the given elements
func assertSeqContents(s Seq, intl []types.Elem, t *testing.T) {
	if ls := ToSlice(s); !intSlicesEq(ls, intl) {
		t.Fatalf("Slice contents wrong: %v not %v", ls, intl)
	}
}

// Asserts that the given Seq has all elements, and only the elements, in the
// given map
func assertSeqContentsNoOrderMap(s Seq, m map[types.Elem]bool, t *testing.T) {
	ls := ToSlice(s)
	if len(ls) != len(m) {
		t.Fatalf("Slice contents wrong: %v not %v", ls, m)
	}
	for i := range ls {
		if _, ok := m[ls[i]]; !ok {
			t.Fatalf("Slice contents wrong: %v not %v", ls, m)
		}
	}
}

// Asserts that the given Seq has all the elements, and only the elements
// (duplicates removed), in the given slice, although no necessarily in the
// order given in the slice
func assertSeqContentsSet(s Seq, ints []types.Elem, t *testing.T) {
	m := map[types.Elem]bool{}
	for i := range ints {
		m[ints[i]] = true
	}
	assertSeqContentsNoOrderMap(s, m, t)
}

func assertSeqContentsHashMap(s Seq, kvs []*KV, t *testing.T) {
	m := map[KV]bool{}
	for i := range kvs {
		m[*kvs[i]] = true
	}
	ls := ToSlice(s)
	if len(ls) != len(m) {
		t.Fatalf("Slice contents wrong: %v not %v", ls, m)
	}
	for i := range ls {
		kv := ls[i].(*KV)
		if _, ok := m[*kv]; !ok {
			t.Fatalf("Slice contents wrong: %v not %v", ls, m)
		}
	}
}

// Asserts that v1 is the same as v2
func assertValue(v1, v2 interface{}, t *testing.T) {
	if gv1, ok := v1.(types.GoType); ok {
		v1 = gv1.V
	}
	if gv2, ok := v2.(types.GoType); ok {
		v2 = gv2.V
	}
	if v1 != v2 {
		t.Logf("Value wrong: %v not %v", v1, v2)
		panic("bail")
	}
}

// Asserts that v1 is a key in the given map
func assertInMap(v1 types.Elem, m map[types.Elem]bool, t *testing.T) {
	if _, ok := m[v1]; !ok {
		t.Fatalf("Value not in set: %v not in %v", v1, m)
	}
}

func keyValV(k, v interface{}) *KV {
	return KeyVal(types.GoType{k}, types.GoType{v})
}
