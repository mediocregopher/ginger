package gg

import (
	"bufio"
	"fmt"
	"io"
)

// Decoder reads Value's off of a byte stream.
type Decoder struct {
	br        *bufio.Reader
	brNextLoc Location

	unread   []locatableRune
	lastRead locatableRune
}

// NewDecoder returns a Decoder which will decode the given stream as a gg
// formatted stream of a Values.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		br:        bufio.NewReader(r),
		brNextLoc: Location{Row: 1, Col: 1},
	}
}

func (d *Decoder) readRune() (locatableRune, error) {
	if len(d.unread) > 0 {
		d.lastRead = d.unread[len(d.unread)-1]
		d.unread = d.unread[:len(d.unread)-1]
		return d.lastRead, nil
	}

	loc := d.brNextLoc

	r, _, err := d.br.ReadRune()
	if err != nil {
		return d.lastRead, err
	}

	if r == '\n' {
		d.brNextLoc.Row++
		d.brNextLoc.Col = 1
	} else {
		d.brNextLoc.Col++
	}

	d.lastRead = locatableRune{loc, r}
	return d.lastRead, nil
}

func (d *Decoder) unreadRune(lr locatableRune) {
	if d.lastRead != lr {
		panic(fmt.Sprintf(
			"unreading rune %#v, but last read rune was %#v", lr, d.lastRead,
		))
	}

	d.unread = append(d.unread, lr)
}

func (d *Decoder) nextLoc() Location {
	if len(d.unread) > 0 {
		return d.unread[len(d.unread)-1].Location
	}

	return d.brNextLoc
}

// Next returns the next top-level value in the stream, or io.EOF.
func (d *Decoder) Next() (Value, error) {
	panic("TODO")
}
