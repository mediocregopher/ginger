package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/buger/goterm"
)

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

type xy [2]int

func (p xy) x() int {
	return p[0]
}

func (p xy) y() int {
	return p[1]
}

func (p xy) add(p2 xy) xy {
	p[0] += p2[0]
	p[1] += p2[1]
	return p
}

////////////////////////////////////////////////////////////////////////////////

type terminal struct {
	cursorPos xy
}

func (t *terminal) moveAbs(to xy) {
	t.cursorPos = to
	goterm.MoveCursor(to.x()+1, to.y()+1)
}

func (t *terminal) size() xy {
	return xy{goterm.Width(), goterm.Height()}
}

////////////////////////////////////////////////////////////////////////////////

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
	pos  xy
	size xy // if unset, auto-determined
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

func (b box) getSize() xy {
	if b.size != (xy{}) {
		return b.size
	}
	var size xy
	for _, line := range b.lines() {
		size[1]++
		if l := len(line); l > size[0] {
			size[0] = l
		}
	}
	return size
}

func (b box) draw(term *terminal) {
	chars := boxDefault
	pos := b.pos
	size := b.getSize()
	w, h := size.x(), size.y()

	// draw top line
	term.moveAbs(pos)
	goterm.Print(chars[boxBorderTL])
	for i := 0; i < w; i++ {
		goterm.Print(chars[boxBorderHoriz])
	}
	goterm.Print(chars[boxBorderTR])

	drawLine := func(line string) {
		pos[1]++
		term.moveAbs(pos)
		goterm.Print(chars[boxBorderVert])
		if len(line) > w {
			line = line[:w]
		}
		goterm.Print(line)
		if b.transparent {
			term.moveAbs(pos.add(xy{w + 1, 0}))
		} else {
			goterm.Print(strings.Repeat(" ", w-len(line)))
		}
		goterm.Print(chars[boxBorderVert])
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
	term.moveAbs(pos)
	goterm.Print(chars[boxBorderBL])
	for i := 0; i < w; i++ {
		goterm.Print(chars[boxBorderHoriz])
	}
	goterm.Print(chars[boxBorderBR])
}

////////////////////////////////////////////////////////////////////////////////

const (
	framerate   = 30
	frameperiod = time.Second / time.Duration(framerate)
)

func debugf(str string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, str, args...)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	{ // exit signal handling, cause ctrl-c doesn't work with goterm otherwise
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go func() {
			<-c
			goterm.Clear()
			goterm.Flush()
			os.Stdout.Sync()
			os.Exit(0)
		}()
	}

	term := new(terminal)

	type movingBox struct {
		box
		xRight bool
		yDown  bool
	}

	randBox := func() movingBox {
		tsize := term.size()
		return movingBox{
			box: box{
				pos: xy{rand.Intn(tsize[0]), rand.Intn(tsize[1])},
			},
			xRight: rand.Intn(1) == 0,
			yDown:  rand.Intn(1) == 0,
		}
	}

	boxes := []movingBox{
		randBox(),
		randBox(),
		randBox(),
		randBox(),
		randBox(),
	}

	for range time.Tick(frameperiod) {
		goterm.Clear()

		termSize := term.size()
		now := time.Now()
		for i := range boxes {
			b := &boxes[i]
			b.body = fmt.Sprintf("%d\n%s", now.Unix(), now.String())

			size := b.getSize()
			if b.pos[0] <= 0 {
				b.xRight = true
			} else if b.pos[0]+size[0]+2 > termSize[0] {
				b.xRight = false
			}
			if b.pos[1] <= 0 {
				b.yDown = true
			} else if b.pos[1]+size[1]+2 > termSize[1] {
				b.yDown = false
			}

			if b.xRight {
				b.pos[0] += 3
			} else {
				b.pos[0] -= 3
			}
			if b.yDown {
				b.pos[1]++
			} else {
				b.pos[1]--
			}

			b.draw(term)
		}
		goterm.Flush()
	}
}
