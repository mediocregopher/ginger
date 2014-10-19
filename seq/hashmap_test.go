package seq

import (
	. "testing"

	"github.com/mediocregopher/ginger/types"
)

func kvints(kvs ...*KV) ([]*KV, []types.Elem) {
	ints := make([]types.Elem, len(kvs))
	for i := range kvs {
		ints[i] = kvs[i]
	}
	return kvs, ints
}

// Test that HashMap implements types.Elem (compile-time check)
func TestHashMapElem(t *T) {
	_ = types.Elem(NewHashMap())
}

// Test creating a Set and calling the Seq interface methods on it
func TestHashMapSeq(t *T) {
	kvs, ints := kvints(
		keyValV(1, "one"),
		keyValV(2, "two"),
	)

	// Testing creation and Seq interface methods
	m := NewHashMap(kvs...)
	ms := testSeqNoOrderGen(t, m, ints)

	// ms should be empty at this point
	assertEmpty(ms, t)
}

// Test that the Equal method on HashMaps works
func TestHashMapEqual(t *T) {
	hm, hm2 := NewHashMap(), NewHashMap()
	assertValue(hm.Equal(hm2), true, t)
	assertValue(hm2.Equal(hm), true, t)

	hm = NewHashMap(keyValV(1, "one"), keyValV(2, "two"))
	assertValue(hm.Equal(hm2), false, t)
	assertValue(hm2.Equal(hm), false, t)
	
	hm2 = NewHashMap(keyValV(1, "one"))
	assertValue(hm.Equal(hm2), false, t)
	assertValue(hm2.Equal(hm), false, t)

	hm2 = NewHashMap(keyValV(1, "one"), keyValV(2, "three?"))
	assertValue(hm.Equal(hm2), false, t)
	assertValue(hm2.Equal(hm), false, t)

	hm2 = NewHashMap(keyValV(1, "one"), keyValV(2, "two"))
	assertValue(hm.Equal(hm2), true, t)
	assertValue(hm2.Equal(hm), true, t)
}

// Test getting values from a HashMap
func TestHashMapGet(t *T) {
	kvs := []*KV{
		keyValV(1, "one"),
		keyValV(2, "two"),
	}

	// Degenerate case
	m := NewHashMap()
	assertEmpty(m, t)
	v, ok := m.Get(types.GoType{1})
	assertValue(v, nil, t)
	assertValue(ok, false, t)

	m = NewHashMap(kvs...)
	v, ok = m.Get(types.GoType{1})
	assertSeqContentsHashMap(m, kvs, t)
	assertValue(v, types.GoType{"one"}, t)
	assertValue(ok, true, t)

	v, ok = m.Get(types.GoType{3})
	assertSeqContentsHashMap(m, kvs, t)
	assertValue(v, nil, t)
	assertValue(ok, false, t)
}

// Test setting values on a HashMap
func TestHashMapSet(t *T) {

	// Set on empty
	m := NewHashMap()
	m1, ok := m.Set(types.GoType{1}, types.GoType{"one"})
	assertEmpty(m, t)
	assertSeqContentsHashMap(m1, []*KV{keyValV(1, "one")}, t)
	assertValue(ok, true, t)

	// Set on same key
	m2, ok := m1.Set(types.GoType{1}, types.GoType{"wat"})
	assertSeqContentsHashMap(m1, []*KV{keyValV(1, "one")}, t)
	assertSeqContentsHashMap(m2, []*KV{keyValV(1, "wat")}, t)
	assertValue(ok, false, t)

	// Set on second new key
	m3, ok := m2.Set(types.GoType{2}, types.GoType{"two"})
	assertSeqContentsHashMap(m2, []*KV{keyValV(1, "wat")}, t)
	assertSeqContentsHashMap(m3, []*KV{keyValV(1, "wat"), keyValV(2, "two")}, t)
	assertValue(ok, true, t)

}

// Test deleting keys from sets
func TestHashMapDel(t *T) {

	kvs := []*KV{
		keyValV(1, "one"),
		keyValV(2, "two"),
		keyValV(3, "three"),
	}
	kvs1 := []*KV{
		keyValV(2, "two"),
		keyValV(3, "three"),
	}

	// Degenerate case
	m := NewHashMap()
	m1, ok := m.Del(types.GoType{1})
	assertEmpty(m, t)
	assertEmpty(m1, t)
	assertValue(ok, false, t)

	// Delete actual key
	m = NewHashMap(kvs...)
	m1, ok = m.Del(types.GoType{1})
	assertSeqContentsHashMap(m, kvs, t)
	assertSeqContentsHashMap(m1, kvs1, t)
	assertValue(ok, true, t)

	// Delete it again!
	m2, ok := m1.Del(types.GoType{1})
	assertSeqContentsHashMap(m1, kvs1, t)
	assertSeqContentsHashMap(m2, kvs1, t)
	assertValue(ok, false, t)

}
