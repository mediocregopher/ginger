package gg

import "fmt"

// Location indicates a position in a stream of bytes identified by column
// within newline-separated rows.
type Location struct {
	Row, Col int
}

func (l Location) errf(str string, args ...any) LocatedError {
	return LocatedError{l, fmt.Errorf(str, args...)}
}

func (l Location) locate() Location { return l }

// LocatedError is an error related to a specific point within a decode gg
// stream.
type LocatedError struct {
	Location
	Err error
}

func (e LocatedError) Error() string {
	return fmt.Sprintf("%d:%d: %v", e.Row, e.Col, e.Err)
}

type locatable interface {
	locate() Location
}

type locatableRune struct {
	Location
	r rune
}

type locatableString struct {
	Location
	str string
}

type locatableSlice[T locatable] []T

func (s locatableSlice[T]) locate() Location {
	if len(s) == 0 {
		panic("can't locate empty locatableSlice")
	}
	return s[0].locate()
}
