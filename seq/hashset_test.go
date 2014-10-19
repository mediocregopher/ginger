package seq

import (
	. "testing"

	"github.com/mediocregopher/ginger/types"
)

// Test that HashSet implements types.Elem (compile-time check)
func TestSetElem(t *T) {
	_ = types.Elem(NewSet())
}

// Test creating a Set and calling the Seq interface methods on it
func TestSetSeq(t *T) {
	ints := elemSliceV(nil, 1, "a", 5.0)

	// Testing creation and Seq interface methods
	s := NewSet(ints...)
	ss := testSeqNoOrderGen(t, s, ints)

	// ss should be empty at this point
	s = ToSet(ss)
	var nilpointer *Set
	assertEmpty(s, t)
	assertValue(s, nilpointer, t)
	assertValue(len(ToSlice(s)), 0, t)
}

// Test that the Equal method on Sets works
func TestSetEqual(t *T) {
	s, s2 := NewSet(), NewSet()
	assertValue(s.Equal(s2), true, t)
	assertValue(s2.Equal(s), true, t)

	s = NewSet(elemSliceV(0, 1, 2)...)
	assertValue(s.Equal(s2), false, t)
	assertValue(s2.Equal(s), false, t)

	s2 = NewSet(elemSliceV(0, 1)...)
	assertValue(s.Equal(s2), false, t)
	assertValue(s2.Equal(s), false, t)

	s2 = NewSet(elemSliceV(0, 1, 3)...)
	assertValue(s.Equal(s2), false, t)
	assertValue(s2.Equal(s), false, t)

	s2 = NewSet(elemSliceV(0, 1, 2)...)
	assertValue(s.Equal(s2), true, t)
	assertValue(s2.Equal(s), true, t)
}

// Test setting a value on a Set
func TestSetVal(t *T) {
	ints := elemSliceV(0, 1, 2, 3, 4)
	ints1 := elemSliceV(0, 1, 2, 3, 4, 5)

	// Degenerate case
	s := NewSet()
	assertEmpty(s, t)
	s, ok := s.SetVal(types.GoType{0})
	assertSeqContentsSet(s, elemSliceV(0), t)
	assertValue(ok, true, t)

	s = NewSet(ints...)
	s1, ok := s.SetVal(types.GoType{5})
	assertSeqContentsSet(s, ints, t)
	assertSeqContentsSet(s1, ints1, t)
	assertValue(ok, true, t)

	s2, ok := s1.SetVal(types.GoType{5})
	assertSeqContentsSet(s1, ints1, t)
	assertSeqContentsSet(s2, ints1, t)
	assertValue(ok, false, t)
}

// Test deleting a value from a Set
func TestDelVal(t *T) {
	ints := elemSliceV(0, 1, 2, 3, 4)
	ints1 := elemSliceV(0, 1, 2, 3)
	ints2 := elemSliceV(1, 2, 3, 4)
	ints3 := elemSliceV(1, 2, 3, 4, 5)

	// Degenerate case
	s := NewSet()
	assertEmpty(s, t)
	s, ok := s.DelVal(types.GoType{0})
	assertEmpty(s, t)
	assertValue(ok, false, t)

	s = NewSet(ints...)
	s1, ok := s.DelVal(types.GoType{4})
	assertSeqContentsSet(s, ints, t)
	assertSeqContentsSet(s1, ints1, t)
	assertValue(ok, true, t)

	s1, ok = s1.DelVal(types.GoType{4})
	assertSeqContentsSet(s1, ints1, t)
	assertValue(ok, false, t)

	// 0 is the value on the root node of s, which is kind of a special case. We
	// want to test deleting it and setting a new value (which should get put on
	// the root node).
	s2, ok := s.DelVal(types.GoType{0})
	assertSeqContentsSet(s, ints, t)
	assertSeqContentsSet(s2, ints2, t)
	assertValue(ok, true, t)

	s2, ok = s2.DelVal(types.GoType{0})
	assertSeqContentsSet(s2, ints2, t)
	assertValue(ok, false, t)

	s3, ok := s2.SetVal(types.GoType{5})
	assertSeqContentsSet(s2, ints2, t)
	assertSeqContentsSet(s3, ints3, t)
	assertValue(ok, true, t)
}

// Test getting values from a Set
func GetVal(t *T) {
	//Degenerate case
	s := NewSet()
	v, ok := s.GetVal(types.GoType{1})
	assertValue(v, nil, t)
	assertValue(ok, false, t)

	s = NewSet(elemSliceV(0, 1, 2, 3, 4)...)
	v, ok = s.GetVal(types.GoType{1})
	assertValue(v, 1, t)
	assertValue(ok, true, t)

	// After delete
	s, _ = s.DelVal(types.GoType{1})
	v, ok = s.GetVal(types.GoType{1})
	assertValue(v, nil, t)
	assertValue(ok, false, t)

	// After set
	s, _ = s.SetVal(types.GoType{1})
	v, ok = s.GetVal(types.GoType{1})
	assertValue(v, 1, t)
	assertValue(ok, true, t)

	// After delete root node
	s, _ = s.DelVal(types.GoType{0})
	v, ok = s.GetVal(types.GoType{0})
	assertValue(v, nil, t)
	assertValue(ok, false, t)

	// After set root node
	s, _ = s.SetVal(types.GoType{5})
	v, ok = s.GetVal(types.GoType{5})
	assertValue(v, 5, t)
	assertValue(ok, true, t)
}

// Test that Size functions properly for all cases
func TestSetSize(t *T) {
	// Degenerate case
	s := NewSet()
	assertValue(s.Size(), uint64(0), t)

	// Initialization case
	s = NewSet(elemSliceV(0, 1, 2)...)
	assertValue(s.Size(), uint64(3), t)

	// Setting (both value not in and a value already in)
	s, _ = s.SetVal(types.GoType{3})
	assertValue(s.Size(), uint64(4), t)
	s, _ = s.SetVal(types.GoType{3})
	assertValue(s.Size(), uint64(4), t)

	// Deleting (both value already in and a value not in)
	s, _ = s.DelVal(types.GoType{3})
	assertValue(s.Size(), uint64(3), t)
	s, _ = s.DelVal(types.GoType{3})
	assertValue(s.Size(), uint64(3), t)

	// Deleting and setting the root node
	s, _ = s.DelVal(types.GoType{0})
	assertValue(s.Size(), uint64(2), t)
	s, _ = s.SetVal(types.GoType{5})
	assertValue(s.Size(), uint64(3), t)

}

// Test that Union functions properly
func TestUnion(t *T) {
	// Degenerate case
	empty := NewSet()
	assertEmpty(empty.Union(empty), t)

	ints1 := elemSliceV(0, 1, 2)
	ints2 := elemSliceV(3, 4, 5)
	intsu := append(ints1, ints2...)
	s1 := NewSet(ints1...)
	s2 := NewSet(ints2...)

	assertSeqContentsSet(s1.Union(empty), ints1, t)
	assertSeqContentsSet(empty.Union(s1), ints1, t)

	su := s1.Union(s2)
	assertSeqContentsSet(s1, ints1, t)
	assertSeqContentsSet(s2, ints2, t)
	assertSeqContentsSet(su, intsu, t)
}

// Test that Intersection functions properly
func TestIntersection(t *T) {
	// Degenerate case
	empty := NewSet()
	assertEmpty(empty.Intersection(empty), t)

	ints1 := elemSliceV(0, 1, 2)
	ints2 := elemSliceV(1, 2, 3)
	ints3 := elemSliceV(4, 5, 6)
	intsi := elemSliceV(1, 2)
	s1 := NewSet(ints1...)
	s2 := NewSet(ints2...)
	s3 := NewSet(ints3...)

	assertEmpty(s1.Intersection(empty), t)
	assertEmpty(empty.Intersection(s1), t)

	si := s1.Intersection(s2)
	assertEmpty(s1.Intersection(s3), t)
	assertSeqContentsSet(s1, ints1, t)
	assertSeqContentsSet(s2, ints2, t)
	assertSeqContentsSet(s3, ints3, t)
	assertSeqContentsSet(si, intsi, t)
}

// Test that Difference functions properly
func TestDifference(t *T) {
	// Degenerate case
	empty := NewSet()
	assertEmpty(empty.Difference(empty), t)

	ints1 := elemSliceV(0, 1, 2, 3)
	ints2 := elemSliceV(2, 3, 4)
	intsd := elemSliceV(0, 1)
	s1 := NewSet(ints1...)
	s2 := NewSet(ints2...)

	assertSeqContentsSet(s1.Difference(empty), ints1, t)
	assertEmpty(empty.Difference(s1), t)

	sd := s1.Difference(s2)
	assertSeqContentsSet(s1, ints1, t)
	assertSeqContentsSet(s2, ints2, t)
	assertSeqContentsSet(sd, intsd, t)
}

// Test that SymDifference functions properly
func TestSymDifference(t *T) {
	// Degenerate case
	empty := NewSet()
	assertEmpty(empty.SymDifference(empty), t)

	ints1 := elemSliceV(0, 1, 2, 3)
	ints2 := elemSliceV(2, 3, 4)
	intsd := elemSliceV(0, 1, 4)
	s1 := NewSet(ints1...)
	s2 := NewSet(ints2...)

	assertSeqContentsSet(s1.SymDifference(empty), ints1, t)
	assertSeqContentsSet(empty.SymDifference(s1), ints1, t)

	sd := s1.SymDifference(s2)
	assertSeqContentsSet(s1, ints1, t)
	assertSeqContentsSet(s2, ints2, t)
	assertSeqContentsSet(sd, intsd, t)
}
