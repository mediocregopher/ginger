package seq

import (
	. "testing"
	"time"

	"github.com/mediocregopher/ginger/types"
)

// Test that Lazy implements types.Elem (compile-time check)
func TestLazyElem(t *T) {
	_ = types.Elem(NewLazy(nil))
}

// Test lazy operation and thread-safety
func TestLazyBasic(t *T) {
	ch := make(chan types.GoType)
	mapfn := func(el types.Elem) types.Elem {
		i := el.(types.GoType)
		ch <- i
		return i
	}

	intl := elemSliceV(0, 1, 2, 3, 4)
	l := NewList(intl...)
	ml := LMap(mapfn, l)

	// ml is a lazy list of intl, which will write to ch the first time any of
	// the elements are read. This for loop ensures ml is thread-safe
	for i := 0; i < 10; i++ {
		go func() {
			mlintl := ToSlice(ml)
			if !intSlicesEq(mlintl, intl) {
				panic("contents not right")
			}
		}()
	}

	// This loop and subsequent close ensure that ml only ever "creates" each
	// element once
	for _, el := range intl {
		select {
		case elch := <-ch:
			assertValue(el, elch, t)
		case <-time.After(1 * time.Millisecond):
			t.Fatalf("Took too long reading result")
		}
	}
	close(ch)
}

// Test that arbitrary Seqs can turn into Lazy
func TestToLazy(t *T) {
	intl := elemSliceV(0, 1, 2, 3, 4)
	l := NewList(intl...)
	ll := ToLazy(l)
	assertSeqContents(ll, intl, t)
}
