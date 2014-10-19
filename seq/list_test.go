package seq

import (
	. "testing"

	"github.com/mediocregopher/ginger/types"
)

// Test that List implements types.Elem (compile-time check)
func TestListElem(t *T) {
	_ = types.Elem(NewList())
}

// Asserts that the given list is properly formed and has all of its size fields
// filled in correctly
func assertSaneList(l *List, t *T) {
	if Size(l) == 0 {
		var nilpointer *List
		assertValue(l, nilpointer, t)
		return
	}

	size := Size(l)
	assertValue(Size(l.next), size-1, t)
	assertSaneList(l.next, t)
}

// Test creating a list and calling the Seq interface methods on it
func TestListSeq(t *T) {
	ints := elemSliceV(1, "a", 5.0)

	// Testing creation and Seq interface methods
	l := NewList(ints...)
	sl := testSeqGen(t, l, ints)

	// sl should be empty at this point
	l = ToList(sl)
	var nilpointer *List
	assertEmpty(l, t)
	assertValue(l, nilpointer, t)
	assertValue(len(ToSlice(l)), 0, t)

	// Testing creation of empty List.
	emptyl := NewList()
	assertValue(emptyl, nilpointer, t)
}

// Test that the Equal method on Lists works
func TestListEqual(t *T) {
	l, l2 := NewList(), NewList()
	assertValue(l.Equal(l2), true, t)
	assertValue(l2.Equal(l), true, t)

	l2 = NewList(elemSliceV(1, 2, 3)...)
	assertValue(l.Equal(l2), false, t)
	assertValue(l2.Equal(l), false, t)

	l = NewList(elemSliceV(1, 2, 3, 4)...)
	assertValue(l.Equal(l2), false, t)
	assertValue(l2.Equal(l), false, t)

	l2 = NewList(elemSliceV(1, 2, 3, 4)...)
	assertValue(l.Equal(l2), true, t)
	assertValue(l2.Equal(l), true, t)
}

// Test the string representation of a List
func TestStringSeq(t *T) {
	l := NewList(elemSliceV(0, 1, 2, 3)...)
	assertValue(l.String(), "( 0 1 2 3 )", t)

	l = NewList(elemSliceV(
		0, 1, 2,
		NewList(elemSliceV(3, 4)...),
		5,
		NewList(elemSliceV(6, 7, 8)...))...)
	assertValue(l.String(), "( 0 1 2 ( 3 4 ) 5 ( 6 7 8 ) )", t)
}

// Test prepending an element to the beginning of a list
func TestPrepend(t *T) {
	// Normal case
	intl := elemSliceV(3, 2, 1, 0)
	l := NewList(intl...)
	nl := l.Prepend(types.GoType{4})
	assertSaneList(l, t)
	assertSaneList(nl, t)
	assertSeqContents(l, intl, t)
	assertSeqContents(nl, elemSliceV(4, 3, 2, 1, 0), t)

	// Degenerate case
	l = NewList()
	nl = l.Prepend(types.GoType{0})
	assertEmpty(l, t)
	assertSaneList(nl, t)
	assertSeqContents(nl, elemSliceV(0), t)
}

// Test prepending a Seq to the beginning of a list
func TestPrependSeq(t *T) {
	//Normal case
	intl1 := elemSliceV(3, 4)
	intl2 := elemSliceV(0, 1, 2)
	l1 := NewList(intl1...)
	l2 := NewList(intl2...)
	nl := l1.PrependSeq(l2)
	assertSaneList(l1, t)
	assertSaneList(l2, t)
	assertSaneList(nl, t)
	assertSeqContents(l1, intl1, t)
	assertSeqContents(l2, intl2, t)
	assertSeqContents(nl, elemSliceV(0, 1, 2, 3, 4), t)

	// Degenerate cases
	blank1 := NewList()
	blank2 := NewList()
	nl = blank1.PrependSeq(blank2)
	assertEmpty(blank1, t)
	assertEmpty(blank2, t)
	assertEmpty(nl, t)

	nl = blank1.PrependSeq(l1)
	assertEmpty(blank1, t)
	assertSaneList(nl, t)
	assertSeqContents(nl, intl1, t)

	nl = l1.PrependSeq(blank1)
	assertEmpty(blank1, t)
	assertSaneList(nl, t)
	assertSeqContents(nl, intl1, t)
}

// Test appending to the end of a List
func TestAppend(t *T) {
	// Normal case
	intl := elemSliceV(3, 2, 1)
	l := NewList(intl...)
	nl := l.Append(types.GoType{0})
	assertSaneList(l, t)
	assertSaneList(nl, t)
	assertSeqContents(l, intl, t)
	assertSeqContents(nl, elemSliceV(3, 2, 1, 0), t)

	// Edge case (algorithm gets weird here)
	l = NewList(elemSliceV(1)...)
	nl = l.Append(types.GoType{0})
	assertSaneList(l, t)
	assertSaneList(nl, t)
	assertSeqContents(l, elemSliceV(1), t)
	assertSeqContents(nl, elemSliceV(1, 0), t)

	// Degenerate case
	l = NewList()
	nl = l.Append(types.GoType{0})
	assertEmpty(l, t)
	assertSaneList(nl, t)
	assertSeqContents(nl, elemSliceV(0), t)
}

// Test retrieving items from a List
func TestNth(t *T) {
	// Normal case, in bounds
	intl := elemSliceV(0, 2, 4, 6, 8)
	l := NewList(intl...)
	r, ok := l.Nth(3)
	assertSaneList(l, t)
	assertSeqContents(l, intl, t)
	assertValue(r, 6, t)
	assertValue(ok, true, t)

	// Normal case, out of bounds
	r, ok = l.Nth(8)
	assertSaneList(l, t)
	assertSeqContents(l, intl, t)
	assertValue(r, nil, t)
	assertValue(ok, false, t)

	// Degenerate case
	l = NewList()
	r, ok = l.Nth(0)
	assertEmpty(l, t)
	assertValue(r, nil, t)
	assertValue(ok, false, t)
}
