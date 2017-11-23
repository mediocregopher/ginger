// Package terminal implements functionality related to interacting with a
// terminal. Using this package takes the place of using stdout directly
package terminal

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"syscall"
	"unicode/utf8"
	"unsafe"

	"github.com/mediocregopher/ginger/gim/geo"
)

// Terminal provides an interface to a terminal which allows for "drawing"
// rather than just writing. Note that all operations on a Terminal aren't
// actually drawn to the screen until Flush is called.
//
// The coordinate system described by Terminal looks like this:
//
// 0,0 ------------------> x
//  |
//  |
//  |
//  |
//  |
//  |
//  |
//  |
//  v
//  y
//
type Terminal struct {
	buf *bytes.Buffer
	pos geo.XY

	// When initialized this will be set to os.Stdout, but can be set to
	// anything
	Out io.Writer
}

// New initializes and returns a usable Terminal
func New() *Terminal {
	return &Terminal{
		buf: new(bytes.Buffer),
		Out: os.Stdout,
	}
}

// WindowSize returns the size of the terminal window (width/height)
// TODO this doesn't support winblows
func (t *Terminal) WindowSize() geo.XY {
	var sz struct {
		rows    uint16
		cols    uint16
		xpixels uint16
		ypixels uint16
	}
	_, _, err := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(&sz)),
	)
	if err != 0 {
		panic(err.Error())
	}
	return geo.XY{int(sz.cols), int(sz.rows)}
}

// MoveCursorTo moves the cursor to the given position
func (t *Terminal) MoveCursorTo(to geo.XY) {
	// actual terminal uses 1,1 as top-left, because 1-indexing is a great idea
	fmt.Fprintf(t.buf, "\033[%d;%dH", to[1]+1, to[0]+1)
	t.pos = to
}

// MoveCursor moves the cursor relative to its current position by the given
// vector
func (t *Terminal) MoveCursor(by geo.XY) {
	t.MoveCursorTo(t.pos.Add(by))
}

// HideCursor causes the cursor to not actually be shown
func (t *Terminal) HideCursor() {
	fmt.Fprintf(t.buf, "\033[?25l")
}

// ShowCursor causes the cursor to be shown, if it was previously hidden
func (t *Terminal) ShowCursor() {
	fmt.Fprintf(t.buf, "\033[?25h")
}

// Reset completely clears all drawn characters on the screen and returns the
// cursor to the origin
func (t *Terminal) Reset() {
	fmt.Fprintf(t.buf, "\033[2J")
}

// Printf prints the given formatted string to the terminal, updating the
// internal cursor position accordingly
func (t *Terminal) Printf(format string, args ...interface{}) {
	str := fmt.Sprintf(format, args...)
	t.buf.WriteString(str)
	t.pos[0] += utf8.RuneCountInString(str)
}

// Flush writes all buffered changes to the screen
func (t *Terminal) Flush() {
	if _, err := io.Copy(t.Out, t.buf); err != nil {
		panic(err)
	}
}

// TODO deal with these

const (
	// Reset all custom styles
	ansiReset = "\033[0m"

	// Reset to default color
	ansiResetColor = "\033[32m"

	// Return curor to start of line and clean it
	ansiResetLine = "\r\033[K"
)

// List of possible colors
const (
	black = iota
	red
	green
	yellow
	blue
	magenta
	cyan
	white
)

func getFgColor(code int) string {
	return fmt.Sprintf("\033[3%dm", code)
}

func getBgColor(code int) string {
	return fmt.Sprintf("\033[4%dm", code)
}

func fgColor(str string, color int) string {
	return fmt.Sprintf("%s%s%s", getFgColor(color), str, ansiReset)
}

func bgColor(str string, color int) string {
	return fmt.Sprintf("%s%s%s", getBgColor(color), str, ansiReset)
}
