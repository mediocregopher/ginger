package terminal

import (
	"fmt"
	"strconv"
	"unicode"

	"github.com/mediocregopher/ginger/gim/geo"
)

// Reset all custom styles
const ansiReset = "\033[0m"

// Color describes the foreground or background color of text
type Color int

// Available Color values
const (
	// whatever the terminal's default color scheme is
	Default = iota

	Black
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
)

type bufStyle struct {
	fgColor Color
	bgColor Color
}

// returns foreground and background ansi codes
func (bf bufStyle) ansi() (string, string) {
	var fg, bg string
	if bf.fgColor != Default {
		fg = "\033[0;3" + strconv.Itoa(int(bf.fgColor)-1) + "m"
	}
	if bf.bgColor != Default {
		bg = "\033[0;4" + strconv.Itoa(int(bf.bgColor)-1) + "m"
	}
	return fg, bg
}

// returns the ansi sequence which would modify the style to the given one
func (bf bufStyle) diffTo(bf2 bufStyle) string {
	// this implementation is naive, but whatever
	if bf == bf2 {
		return ""
	}

	fg, bg := bf2.ansi()
	if (bf == bufStyle{}) {
		return fg + bg
	}
	return ansiReset + fg + bg
}

type bufPoint struct {
	r rune
	bufStyle
}

// Buffer describes an infinitely sized terminal buffer to which anything may be
// drawn, and which will efficiently generate strings representing the drawn
// text.
type Buffer struct {
	currStyle bufStyle
	currPos   geo.XY
	m         *mat
	max       geo.XY
}

// NewBuffer initializes and returns a new empty buffer. The proper way to clear
// a buffer is to toss the old one and generate a new one.
func NewBuffer() *Buffer {
	return &Buffer{
		m:   newMat(),
		max: geo.XY{-1, -1},
	}
}

// Copy creates a new identical instance of this Buffer and returns it.
func (b *Buffer) Copy() *Buffer {
	b2 := NewBuffer()
	b.m.iter(func(x, y int, v interface{}) bool {
		b2.setRune(geo.XY{x, y}, v.(bufPoint))
		return true
	})
	b2.currStyle = b.currStyle
	b2.currPos = b.currPos
	return b2
}

func (b *Buffer) setRune(at geo.XY, p bufPoint) {
	b.m.set(at[0], at[1], p)
	b.max = b.max.Max(at)
}

// WriteRune writes the given rune to the Buffer at whatever the current
// position is, with whatever the current styling is.
func (b *Buffer) WriteRune(r rune) {
	if r == '\n' {
		b.currPos[0], b.currPos[1] = 0, b.currPos[1]+1
		return
	} else if r == '\r' {
		b.currPos[0] = 0
	} else if !unicode.IsPrint(r) {
		panic(fmt.Sprintf("character %q is not supported by terminal.Buffer", r))
	}

	b.setRune(b.currPos, bufPoint{
		r:        r,
		bufStyle: b.currStyle,
	})
	b.currPos[0]++
}

// WriteString writes the given string to the Buffer at whatever the current
// position is, with whatever the current styling is.
func (b *Buffer) WriteString(s string) {
	for _, r := range s {
		b.WriteRune(r)
	}
}

// SetPos sets the cursor position in the Buffer, so Print operations will begin
// at that point. Remember that the origin is at point (0, 0).
func (b *Buffer) SetPos(xy geo.XY) {
	b.currPos = xy
}

// SetFGColor sets subsequent text's foreground color.
func (b *Buffer) SetFGColor(c Color) {
	b.currStyle.fgColor = c
}

// SetBGColor sets subsequent text's background color.
func (b *Buffer) SetBGColor(c Color) {
	b.currStyle.bgColor = c
}

// ResetStyle unsets all text styling options which have been set.
func (b *Buffer) ResetStyle() {
	b.currStyle = bufStyle{}
}

// String renders and returns a string which, when printed to a terminal, will
// print the Buffer's contents at the terminal's current cursor position.
func (b *Buffer) String() string {
	s := ansiReset // always start with a reset
	var style bufStyle
	var pos geo.XY
	move := func(to geo.XY) {
		diff := to.Sub(pos)
		if diff[0] > 0 {
			s += "\033[" + strconv.Itoa(diff[0]) + "C"
		} else if diff[0] < 0 {
			s += "\033[" + strconv.Itoa(-diff[0]) + "D"
		}
		if diff[1] > 0 {
			s += "\033[" + strconv.Itoa(diff[1]) + "B"
		} else if diff[1] < 0 {
			s += "\033[" + strconv.Itoa(-diff[1]) + "A"
		}
		pos = to
	}

	b.m.iter(func(x, y int, v interface{}) bool {
		p := v.(bufPoint)
		move(geo.XY{x, y})
		s += style.diffTo(p.bufStyle)
		style = p.bufStyle
		s += string(p.r)
		pos[0]++
		return true
	})
	return s
}

// DrawBuffer copies the given Buffer onto this one, with the given's top-left
// corner being at the given position. The given buffer may be the same as this
// one.
//
// Calling this method does not affect this Buffer's current cursor position or
// style.
func (b *Buffer) DrawBuffer(at geo.XY, b2 *Buffer) {
	if b == b2 {
		b2 = b2.Copy()
	}
	b2.m.iter(func(x, y int, v interface{}) bool {
		x += at[0]
		y += at[1]
		if x < 0 || y < 0 {
			return true
		}
		b.setRune(geo.XY{x, y}, v.(bufPoint))
		return true
	})
}

// DrawBufferCentered is like DrawBuffer, but centered around the given point
// instead of translated by it.
func (b *Buffer) DrawBufferCentered(around geo.XY, b2 *Buffer) {
	b2rect := geo.Rect{Size: b2.Size()}
	b.DrawBuffer(b2rect.Centered(around).TopLeft, b2)
}

// Size returns the dimensions of the Buffer's current area which has been
// written to.
func (b *Buffer) Size() geo.XY {
	return b.max.Add(geo.XY{1, 1})
}
