// Package terminal implements functionality related to interacting with a
// terminal. Using this package takes the place of using stdout directly
package terminal

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"syscall"
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

// SetPos sets the terminal's actual cursor position to the given coordinates.
func (t *Terminal) SetPos(to geo.XY) {
	// actual terminal uses 1,1 as top-left, because 1-indexing is a great idea
	fmt.Fprintf(t.buf, "\033[%d;%dH", to[1]+1, to[0]+1)
}

// HideCursor causes the cursor to not actually be shown
func (t *Terminal) HideCursor() {
	fmt.Fprintf(t.buf, "\033[?25l")
}

// ShowCursor causes the cursor to be shown, if it was previously hidden
func (t *Terminal) ShowCursor() {
	fmt.Fprintf(t.buf, "\033[?25h")
}

// Clear completely clears all drawn characters on the screen and returns the
// cursor to the origin. This implicitly calls Draw.
func (t *Terminal) Clear() {
	t.buf.Reset()
	fmt.Fprintf(t.buf, "\033[2J")
	t.Draw()
}

// WriteBuffer writes the contents to the Buffer to the Terminal's buffer,
// starting at the given coordinate.
func (t *Terminal) WriteBuffer(at geo.XY, b *Buffer) {
	t.SetPos(at)
	t.buf.WriteString(b.String())
}

// Draw writes all buffered changes to the screen
func (t *Terminal) Draw() {
	if _, err := io.Copy(t.Out, t.buf); err != nil {
		panic(err)
	}
	t.buf.Reset()
}
