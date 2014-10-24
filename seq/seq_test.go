package seq

import (
	. "testing"

	"github.com/mediocregopher/ginger/types"
)

// Tests the FirstRest, Size, Empty, and ToSlice methods of a Seq
func testSeqGen(t *T, s Seq, ints []types.Elem) Seq {
	intsl := uint64(len(ints))
	for i := range ints {
		assertSaneList(ToList(s), t)
		assertValue(Size(s), intsl-uint64(i), t)
		assertSeqContents(s, ints[i:], t)

		first, rest, ok := s.FirstRest()
		assertValue(ok, true, t)
		assertValue(first, ints[i], t)

		empty := Empty(s)
		assertValue(empty, false, t)

		s = rest
	}
	empty := Empty(s)
	assertValue(empty, true, t)
	return s
}

// Tests the FirstRest, Size, and ToSlice methods of an unordered Seq
func testSeqNoOrderGen(t *T, s Seq, ints []types.Elem) Seq {
	intsl := uint64(len(ints))

	m := map[types.Elem]bool{}
	for i := range ints {
		m[ints[i]] = true
	}

	for i := range ints {
		assertSaneList(ToList(s), t)
		assertValue(Size(s), intsl-uint64(i), t)
		assertSeqContentsNoOrderMap(s, m, t)

		first, rest, ok := s.FirstRest()
		assertValue(ok, true, t)
		assertInMap(first, m, t)

		delete(m, first)
		s = rest
	}
	return s
}

// Test reversing a Seq
func TestReverse(t *T) {
	// Normal case
	intl := elemSliceV(3, 2, 1)
	l := NewList(intl...)
	nl := Reverse(l)
	assertSeqContents(l, intl, t)
	assertSeqContents(nl, elemSliceV(1, 2, 3), t)

	// Degenerate case
	l = NewList()
	nl = Reverse(l)
	assertEmpty(l, t)
	assertEmpty(nl, t)
}

func testMapGen(t *T, mapFn func(func(types.Elem) types.Elem, Seq) Seq) {
	fn := func(n types.Elem) types.Elem {
		return types.GoType{n.(types.GoType).V.(int) + 1}
	}

	// Normal case
	intl := elemSliceV(1, 2, 3)
	l := NewList(intl...)
	nl := mapFn(fn, l)
	assertSeqContents(l, intl, t)
	assertSeqContents(nl, elemSliceV(2, 3, 4), t)

	// Degenerate case
	l = NewList()
	nl = mapFn(fn, l)
	assertEmpty(l, t)
	assertEmpty(nl, t)
}

// Test mapping over a Seq
func TestMap(t *T) {
	testMapGen(t, Map)
}

// Test lazily mapping over a Seq
func TestLMap(t *T) {
	testMapGen(t, LMap)
}

// Test reducing over a Seq
func TestReduce(t *T) {
	fn := func(acc, el types.Elem) (types.Elem, bool) {
		acci := acc.(types.GoType).V.(int)
		eli := el.(types.GoType).V.(int)
		return types.GoType{acci + eli}, false
	}

	// Normal case
	intl := elemSliceV(1, 2, 3, 4)
	l := NewList(intl...)
	r := Reduce(fn, types.GoType{0}, l)
	assertSeqContents(l, intl, t)
	assertValue(r, 10, t)

	// Short-circuit case
	fns := func(acc, el types.Elem) (types.Elem, bool) {
		acci := acc.(types.GoType).V.(int)
		eli := el.(types.GoType).V.(int)
		return types.GoType{acci + eli}, eli > 2
	}
	r = Reduce(fns, types.GoType{0}, l)
	assertSeqContents(l, intl, t)
	assertValue(r, 6, t)

	// Degenerate case
	l = NewList()
	r = Reduce(fn, types.GoType{0}, l)
	assertEmpty(l, t)
	assertValue(r, 0, t)
}

// Test the Any function
func TestAny(t *T) {
	fn := func(el types.Elem) bool {
		return el.(types.GoType).V.(int) > 3
	}

	// Value found case
	intl := elemSliceV(1, 2, 3, 4)
	l := NewList(intl...)
	r, ok := Any(fn, l)
	assertSeqContents(l, intl, t)
	assertValue(r, 4, t)
	assertValue(ok, true, t)

	// Value not found case
	intl = elemSliceV(1, 2, 3)
	l = NewList(intl...)
	r, ok = Any(fn, l)
	assertSeqContents(l, intl, t)
	assertValue(r, nil, t)
	assertValue(ok, false, t)

	// Degenerate case
	l = NewList()
	r, ok = Any(fn, l)
	assertEmpty(l, t)
	assertValue(r, nil, t)
	assertValue(ok, false, t)
}

// Test the All function
func TestAll(t *T) {
	fn := func(el types.Elem) bool {
		return el.(types.GoType).V.(int) > 3
	}

	// All match case
	intl := elemSliceV(4, 5, 6)
	l := NewList(intl...)
	ok := All(fn, l)
	assertSeqContents(l, intl, t)
	assertValue(ok, true, t)

	// Not all match case
	intl = elemSliceV(3, 4, 2, 5)
	l = NewList(intl...)
	ok = All(fn, l)
	assertSeqContents(l, intl, t)
	assertValue(ok, false, t)

	// Degenerate case
	l = NewList()
	ok = All(fn, l)
	assertEmpty(l, t)
	assertValue(ok, true, t)
}

func testFilterGen(t *T, filterFn func(func(types.Elem) bool, Seq) Seq) {
	fn := func(el types.Elem) bool {
		return el.(types.GoType).V.(int)%2 != 0
	}

	// Normal case
	intl := elemSliceV(1, 2, 3, 4, 5)
	l := NewList(intl...)
	r := filterFn(fn, l)
	assertSeqContents(l, intl, t)
	assertSeqContents(r, elemSliceV(1, 3, 5), t)

	// Degenerate cases
	l = NewList()
	r = filterFn(fn, l)
	assertEmpty(l, t)
	assertEmpty(r, t)
}

// Test the Filter function
func TestFilter(t *T) {
	testFilterGen(t, Filter)
}

// Test the lazy Filter function
func TestLFilter(t *T) {
	testFilterGen(t, LFilter)
}

// Test Flatten-ing of a Seq
func TestFlatten(t *T) {
	// Normal case
	intl1 := elemSliceV(0, 1, 2)
	intl2 := elemSliceV(3, 4, 5)
	l1 := NewList(intl1...)
	l2 := NewList(intl2...)
	blank := NewList()
	intl := elemSliceV(-1, l1, l2, 6, blank, 7)
	l := NewList(intl...)
	nl := Flatten(l)
	assertSeqContents(l1, intl1, t)
	assertSeqContents(l2, intl2, t)
	assertEmpty(blank, t)
	assertSeqContents(l, intl, t)
	assertSeqContents(nl, elemSliceV(-1, 0, 1, 2, 3, 4, 5, 6, 7), t)

	// Degenerate case
	nl = Flatten(blank)
	assertEmpty(blank, t)
	assertEmpty(nl, t)
}

func testTakeGen(t *T, takeFn func(uint64, Seq) Seq) {
	// Normal case
	intl := elemSliceV(0, 1, 2, 3, 4)
	l := NewList(intl...)
	nl := takeFn(3, l)
	assertSeqContents(l, intl, t)
	assertSeqContents(nl, elemSliceV(0, 1, 2), t)

	// Edge cases
	nl = takeFn(5, l)
	assertSeqContents(l, intl, t)
	assertSeqContents(nl, intl, t)

	nl = takeFn(6, l)
	assertSeqContents(l, intl, t)
	assertSeqContents(nl, intl, t)

	// Degenerate cases
	empty := NewList()
	nl = takeFn(1, empty)
	assertEmpty(empty, t)
	assertEmpty(nl, t)

	nl = takeFn(0, l)
	assertSeqContents(l, intl, t)
	assertEmpty(nl, t)
}

// Test taking from a Seq
func TestTake(t *T) {
	testTakeGen(t, Take)
}

// Test lazily taking from a Seq
func TestLTake(t *T) {
	testTakeGen(t, LTake)
}

func testTakeWhileGen(t *T, takeWhileFn func(func(types.Elem) bool, Seq) Seq) {
	pred := func(el types.Elem) bool {
		return el.(types.GoType).V.(int) < 3
	}

	// Normal case
	intl := elemSliceV(0, 1, 2, 3, 4, 5)
	l := NewList(intl...)
	nl := takeWhileFn(pred, l)
	assertSeqContents(l, intl, t)
	assertSeqContents(nl, elemSliceV(0, 1, 2), t)

	// Edge cases
	intl = elemSliceV(5, 5, 5)
	l = NewList(intl...)
	nl = takeWhileFn(pred, l)
	assertSeqContents(l, intl, t)
	assertEmpty(nl, t)

	intl = elemSliceV(0, 1, 2)
	l = NewList(intl...)
	nl = takeWhileFn(pred, l)
	assertSeqContents(l, intl, t)
	assertSeqContents(nl, elemSliceV(0, 1, 2), t)

	// Degenerate case
	l = NewList()
	nl = takeWhileFn(pred, l)
	assertEmpty(l, t)
	assertEmpty(nl, t)
}

// Test taking from a Seq until a given condition
func TestTakeWhile(t *T) {
	testTakeWhileGen(t, TakeWhile)
}

// Test lazily taking from a Seq until a given condition
func TestLTakeWhile(t *T) {
	testTakeWhileGen(t, LTakeWhile)
}

// Test dropping from a Seq
func TestDrop(t *T) {
	// Normal case
	intl := elemSliceV(0, 1, 2, 3, 4)
	l := NewList(intl...)
	nl := Drop(3, l)
	assertSeqContents(l, intl, t)
	assertSeqContents(nl, elemSliceV(3, 4), t)

	// Edge cases
	nl = Drop(5, l)
	assertSeqContents(l, intl, t)
	assertEmpty(nl, t)

	nl = Drop(6, l)
	assertSeqContents(l, intl, t)
	assertEmpty(nl, t)

	// Degenerate cases
	empty := NewList()
	nl = Drop(1, empty)
	assertEmpty(empty, t)
	assertEmpty(nl, t)

	nl = Drop(0, l)
	assertSeqContents(l, intl, t)
	assertSeqContents(nl, intl, t)
}

// Test dropping from a Seq until a given condition
func TestDropWhile(t *T) {
	pred := func(el types.Elem) bool {
		return el.(types.GoType).V.(int) < 3
	}

	// Normal case
	intl := elemSliceV(0, 1, 2, 3, 4, 5)
	l := NewList(intl...)
	nl := DropWhile(pred, l)
	assertSeqContents(l, intl, t)
	assertSeqContents(nl, elemSliceV(3, 4, 5), t)

	// Edge cases
	intl = elemSliceV(5, 5, 5)
	l = NewList(intl...)
	nl = DropWhile(pred, l)
	assertSeqContents(l, intl, t)
	assertSeqContents(nl, intl, t)

	intl = elemSliceV(0, 1, 2)
	l = NewList(intl...)
	nl = DropWhile(pred, l)
	assertSeqContents(l, intl, t)
	assertEmpty(nl, t)

	// Degenerate case
	l = NewList()
	nl = DropWhile(pred, l)
	assertEmpty(l, t)
	assertEmpty(nl, t)
}

// Test Traversing a Seq until a given condition
func TestTraverse(t *T) {
	var acc int
	pred := func(el types.Elem) bool {
		acc += el.(types.GoType).V.(int)
		return true
	}

	l := NewList()
	acc = 0
	Traverse(pred, l)
	assertValue(acc, 0, t)

	l2 := NewList(elemSliceV(0, 1, 2, 3)...)
	acc = 0
	Traverse(pred, l2)
	assertValue(acc, 6, t)

	l3 := NewList(
		types.GoType{1},
		types.GoType{2},
		NewList(elemSliceV(4, 5, 6)...),
		types.GoType{3},
	)
	acc = 0
	Traverse(pred, l3)
	assertValue(acc, 21, t)

	pred = func(el types.Elem) bool {
		i := el.(types.GoType).V.(int)
		if i > 4 {
			return false
		}
		acc += i
		return true
	}

	acc = 0
	Traverse(pred, l2)
	assertValue(acc, 6, t)

	acc = 0
	Traverse(pred, l3)
	assertValue(acc, 7, t)
}
