package seq

import (
	"github.com/mediocregopher/ginger/types"
)

// A List is an implementation of Seq in the form of a single-linked-list, and
// is used as the underlying structure for Seqs for most methods that return a
// Seq. It is probably the most efficient and simplest of the implementations.
// Even though, conceptually, all Seq operations return a new Seq, the old Seq
// can actually share nodes with the new Seq (if both are Lists), thereby saving
// memory and copies.
type List struct {
	el   types.Elem
	next *List
}

// Returns a new List comprised of the given elements (or no elements, for an
// empty list)
func NewList(els ...types.Elem) *List {
	elsl := len(els)
	if elsl == 0 {
		return nil
	}

	var cur *List
	for i := 0; i < elsl; i++ {
		cur = &List{els[elsl-i-1], cur}
	}
	return cur
}

// Implementation of FirstRest for Seq interface. Completes in O(1) time.
func (l *List) FirstRest() (types.Elem, Seq, bool) {
	if l == nil {
		return nil, l, false
	} else {
		return l.el, l.next, true
	}
}

// Implementation of Equal for types.Elem interface. Completes in O(N) time if e
// is another List.
func (l *List) Equal(e types.Elem) bool {
	l2, ok := e.(*List)
	if !ok {
		return false
	}

	var el, el2 types.Elem
	var ok2 bool

	s, s2 := Seq(l), Seq(l2)

	for {
		el, s, ok = s.FirstRest()
		el2, s2, ok2 = s2.FirstRest()

		if !ok && !ok2 {
			return true
		}
		if ok != ok2 {
			return false
		}

		if !el.Equal(el2) {
			return false
		}
	}
}

// Implementation of String for Stringer interface.
func (l *List) String() string {
	return ToString(l, "(", ")")
}

// Prepends the given element to the front of the list, returning a copy of the
// new list. Completes in O(1) time.
func (l *List) Prepend(el types.Elem) *List {
	return &List{el, l}
}

// Prepends the argument Seq to the beginning of the callee List, returning a
// copy of the new List. Completes in O(N) time, N being the length of the
// argument Seq
func (l *List) PrependSeq(s Seq) *List {
	var first, cur, prev *List
	var el types.Elem
	var ok bool
	for {
		el, s, ok = s.FirstRest()
		if !ok {
			break
		}
		cur = &List{el, nil}
		if first == nil {
			first = cur
		}
		if prev != nil {
			prev.next = cur
		}
		prev = cur
	}

	// prev will be nil if s is empty
	if prev == nil {
		return l
	}

	prev.next = l
	return first
}

// Appends the given element to the end of the List, returning a copy of the new
// List. While most methods on List don't actually copy much data, this one
// copies the entire list. Completes in O(N) time.
func (l *List) Append(el types.Elem) *List {
	var first, cur, prev *List
	for l != nil {
		cur = &List{l.el, nil}
		if first == nil {
			first = cur
		}
		if prev != nil {
			prev.next = cur
		}
		prev = cur
		l = l.next
	}
	final := &List{el, nil}
	if prev == nil {
		return final
	}
	prev.next = final
	return first
}

// Returns the nth index element (starting at 0), with bool being false if i is
// out of bounds. Completes in O(N) time.
func (l *List) Nth(n uint64) (types.Elem, bool) {
	var el types.Elem
	var ok bool
	s := Seq(l)
	for i := uint64(0); ; i++ {
		el, s, ok = s.FirstRest()
		if !ok {
			return nil, false
		} else if i == n {
			return el, true
		}
	}
}

// Returns the elements in the Seq as a List. Has similar properties as
// ToSlice. In general this completes in O(N) time. If the given Seq is already
// a List it will complete in O(1) time.
func ToList(s Seq) *List {
	var ok bool
	var l *List
	if l, ok = s.(*List); ok {
		return l
	}

	var el types.Elem
	for ret := NewList(); ; {
		if el, s, ok = s.FirstRest(); ok {
			ret = ret.Prepend(el)
		} else {
			return Reverse(ret).(*List)
		}
	}
}
