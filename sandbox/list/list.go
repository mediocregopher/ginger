package list

import "fmt"

/*
	+ size isn't really _necessary_ unless O(1) Len is wanted
	+ append doesn't work well on stack
*/

type List struct {
	// in practice this would be a constant size, with the compiler knowing the
	// size
	underlying []int
	head, size int
}

func New(ii ...int) List {
	l := List{
		underlying: make([]int, ii),
		size:       len(ii),
	}
	copy(l.underlying, ii)
	return l
}

func (l List) Len() int {
	return l.size
}

func (l List) HeadTail() (int, List) {
	if l.size == 0 {
		panic(fmt.Sprintf("can't take HeadTail of empty list"))
	}
	return l.underlying[l.head], List{
		underlying: l.underlying,
		head:       l.head + 1,
		size:       l.size - 1,
	}
}
