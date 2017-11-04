package main

import (
	"fmt"
	"strings"

	"github.com/mediocregopher/ginger/gim/geo"
	"github.com/mediocregopher/ginger/gim/terminal"
)

const (
	boxBorderHoriz = iota
	boxBorderVert
	boxBorderTL
	boxBorderTR
	boxBorderBL
	boxBorderBR
)

var boxDefault = []string{
	"─",
	"│",
	"┌",
	"┐",
	"└",
	"┘",
}

type box struct {
	pos  geo.XY
	size geo.XY // if unset, auto-determined
	body string

	transparent bool
}

func (b box) lines() []string {
	lines := strings.Split(b.body, "\n")
	// if the last line is empty don't include it, it means there was a trailing
	// newline (or the whole string is empty)
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func (b box) innerSize() geo.XY {
	if b.size != (geo.XY{}) {
		return b.size
	}
	var size geo.XY
	for _, line := range b.lines() {
		size[1]++
		if l := len(line); l > size[0] {
			size[0] = l
		}
	}
	return size
}

func (b box) rectSize() geo.XY {
	return b.innerSize().Add(geo.XY{2, 2})
}

// edge returns the coordinate of the edge indicated by the given direction (Up,
// Down, Left, or Right). The coordinate will be for the axis applicable to the
// direction, so for Left/Right it will be the x coordinate and for Up/Down the
// y.
func (b box) rectEdge(dir geo.XY) int {
	size := b.rectSize()
	switch dir {
	case geo.Up:
		return b.pos[1]
	case geo.Down:
		return b.pos[1] + size[1]
	case geo.Left:
		return b.pos[0]
	case geo.Right:
		return b.pos[0] + size[0]
	default:
		panic(fmt.Sprintf("unsupported direction: %#v", dir))
	}
}

func (b box) rectCorner(xDir, yDir geo.XY) geo.XY {
	switch {
	case xDir == geo.Left && yDir == geo.Up:
		return b.pos
	case xDir == geo.Right && yDir == geo.Up:
		size := b.rectSize()
		return b.pos.Add(size.Mul(geo.Right)).Add(geo.XY{-1, 0})
	case xDir == geo.Left && yDir == geo.Down:
		size := b.rectSize()
		return b.pos.Add(size.Mul(geo.Down)).Add(geo.XY{0, -1})
	case xDir == geo.Right && yDir == geo.Down:
		size := b.rectSize()
		return b.pos.Add(size).Add(geo.XY{-1, -1})
	default:
		panic(fmt.Sprintf("unsupported rectCorner args: %v, %v", xDir, yDir))
	}
}

func (b box) draw(term *terminal.Terminal) {
	chars := boxDefault
	pos := b.pos
	size := b.innerSize()
	w, h := size[0], size[1]

	// draw top line
	term.MoveCursorTo(pos)
	term.Printf(chars[boxBorderTL])
	for i := 0; i < w; i++ {
		term.Printf(chars[boxBorderHoriz])
	}
	term.Printf(chars[boxBorderTR])

	drawLine := func(line string) {
		pos[1]++
		term.MoveCursorTo(pos)
		term.Printf(chars[boxBorderVert])
		if len(line) > w {
			line = line[:w]
		}
		term.Printf(line)
		if b.transparent {
			term.MoveCursor(geo.XY{w + 1, 0})
		} else {
			term.Printf(strings.Repeat(" ", w-len(line)))
		}
		term.Printf(chars[boxBorderVert])
	}

	// truncate lines if necessary
	lines := b.lines()
	if len(lines) > h {
		lines = lines[:h]
	}

	// draw body
	for _, line := range lines {
		drawLine(line)
	}

	// draw empty lines
	for i := 0; i < h-len(lines); i++ {
		drawLine("")
	}

	// draw bottom line
	pos[1]++
	term.MoveCursorTo(pos)
	term.Printf(chars[boxBorderBL])
	for i := 0; i < w; i++ {
		term.Printf(chars[boxBorderHoriz])
	}
	term.Printf(chars[boxBorderBR])
}
