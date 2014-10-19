package seq

import (
	"github.com/mediocregopher/ginger/types"
)

// A Lazy is an implementation of a Seq which only actually evaluates its
// contents as those contents become needed. Lazys can be chained together, so
// if you have three steps in a pipeline there aren't two intermediate Seqs
// created, only the final resulting one. Lazys are also thread-safe, so
// multiple routines can interact with the same Lazy pointer at the same time
// but the contents will only be evalutated once.
type Lazy struct {
	this types.Elem
	next *Lazy
	ok   bool
	ch   chan struct{}
}

// Given a Thunk, returns a Lazy around that Thunk.
func NewLazy(t Thunk) *Lazy {
	l := &Lazy{ch: make(chan struct{})}
	go func() {
		l.ch <- struct{}{}
		el, next, ok := t()
		l.this = el
		l.next = NewLazy(next)
		l.ok = ok
		close(l.ch)
	}()
	return l
}

// Implementation of FirstRest for Seq interface. Completes in O(1) time.
func (l *Lazy) FirstRest() (types.Elem, Seq, bool) {
	if l == nil {
		return nil, l, false
	}

	// Reading from the channel tells the Lazy to populate the data and prepare
	// the next item in the seq, it closes the channel when it's done that.
	if _, ok := <-l.ch; ok {
		<-l.ch
	}

	if l.ok {
		return l.this, l.next, true
	} else {
		return nil, nil, false
	}
}

// Implementation of Equal for types.Elem interface. Treats a List as another
// Lazy. Completes in O(N) time if e is another List or List.
func (l *Lazy) Equal(e types.Elem) bool {
	var ls2 *List
	if l2, ok := e.(*Lazy); ok {
		ls2 = ToList(l2)
	} else if ls2, ok = e.(*List); ok {
	} else {
		return false
	}
	ls := ToList(l)
	return ls.Equal(ls2)
}

// Implementation of String for Stringer
func (l *Lazy) String() string {
	return ToString(l, "<<", ">>")
}

// Thunks are the building blocks a Lazy. A Thunk returns an element, another
// Thunk, and a boolean representing if the call yielded any results or if it
// was actually empty (true indicates it yielded results).
type Thunk func() (types.Elem, Thunk, bool)

func mapThunk(fn func(types.Elem) types.Elem, s Seq) Thunk {
	return func() (types.Elem, Thunk, bool) {
		el, ns, ok := s.FirstRest()
		if !ok {
			return nil, nil, false
		}

		return fn(el), mapThunk(fn, ns), true
	}
}

// Lazy implementation of Map
func LMap(fn func(types.Elem) types.Elem, s Seq) Seq {
	return NewLazy(mapThunk(fn, s))
}

func filterThunk(fn func(types.Elem) bool, s Seq) Thunk {
	return func() (types.Elem, Thunk, bool) {
		for {
			el, ns, ok := s.FirstRest()
			if !ok {
				return nil, nil, false
			}

			if keep := fn(el); keep {
				return el, filterThunk(fn, ns), true
			} else {
				s = ns
			}
		}
	}
}

// Lazy implementation of Filter
func LFilter(fn func(types.Elem) bool, s Seq) Seq {
	return NewLazy(filterThunk(fn, s))
}

func takeThunk(n uint64, s Seq) Thunk {
	return func() (types.Elem, Thunk, bool) {
		el, ns, ok := s.FirstRest()
		if !ok || n == 0 {
			return nil, nil, false
		}
		return el, takeThunk(n-1, ns), true
	}
}

// Lazy implementation of Take
func LTake(n uint64, s Seq) Seq {
	return NewLazy(takeThunk(n, s))
}

func takeWhileThunk(fn func(types.Elem) bool, s Seq) Thunk {
	return func() (types.Elem, Thunk, bool) {
		el, ns, ok := s.FirstRest()
		if !ok || !fn(el) {
			return nil, nil, false
		}
		return el, takeWhileThunk(fn, ns), true
	}
}

// Lazy implementation of TakeWhile
func LTakeWhile(fn func(types.Elem) bool, s Seq) Seq {
	return NewLazy(takeWhileThunk(fn, s))
}

func toLazyThunk(s Seq) Thunk {
	return func() (types.Elem, Thunk, bool) {
		el, ns, ok := s.FirstRest()
		if !ok {
			return nil, nil, false
		}
		return el, toLazyThunk(ns), true
	}
}

// Returns the Seq as a Lazy. Pointless for linked-lists, but possibly useful
// for other implementations where FirstRest might be costly and the same Seq
// needs to be iterated over many times.
func ToLazy(s Seq) *Lazy {
	return NewLazy(toLazyThunk(s))
}
