package seq

import (
	. "testing"
	"time"

	"github.com/mediocregopher/ginger/types"
)

// Test lazy operation and thread-safety
func TestLazyBasic(t *T) {
	ch := make(chan int)
	mapfn := func(el types.Elem) types.Elem {
		i := el.(int)
		ch <- i
		return i
	}

	intl := []types.Elem{0, 1, 2, 3, 4}
	l := NewList(intl...)
	ml := LMap(mapfn, l)

	for i := 0; i < 10; i++ {
		go func() {
			mlintl := ToSlice(ml)
			if !intSlicesEq(mlintl, intl) {
				panic("contents not right")
			}
		}()
	}

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
	intl := []types.Elem{0, 1, 2, 3, 4}
	l := NewList(intl...)
	ll := ToLazy(l)
	assertSeqContents(ll, intl, t)
}
