package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/mediocregopher/ginger/gim/geo"
	"github.com/mediocregopher/ginger/gim/terminal"
)

const (
	framerate   = 10
	frameperiod = time.Second / time.Duration(framerate)
)

func debugf(str string, args ...interface{}) {
	if !strings.HasSuffix(str, "\n") {
		str += "\n"
	}
	fmt.Fprintf(os.Stderr, str, args...)
}

// TODO
// * Use actual gg graphs and not fake "boxes"
//   - This will involve wrapping the vertices in some way, to preserve position
// * Once gg graphs are used we can use that birds-eye-view to make better
//   decisions about edge placement

func main() {
	rand.Seed(time.Now().UnixNano())
	term := terminal.New()

	type movingBox struct {
		box
		xRight bool
		yDown  bool
	}

	randBox := func() movingBox {
		tsize := term.WindowSize()
		return movingBox{
			box: box{
				pos:  geo.XY{rand.Intn(tsize[0]), rand.Intn(tsize[1])},
				size: geo.XY{30, 2},
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

		// update phase
		termSize := term.WindowSize()
		for i := range boxes {
			b := &boxes[i]
			b.body = fmt.Sprintf("%d) %v", i, b.rectCorner(geo.Left, geo.Up))
			b.body += fmt.Sprintf(" | %v\n", b.rectCorner(geo.Right, geo.Up))
			b.body += fmt.Sprintf("   %v", b.rectCorner(geo.Left, geo.Down))
			b.body += fmt.Sprintf(" | %v", b.rectCorner(geo.Right, geo.Down))

			size := b.rectSize()
			if b.pos[0] <= 0 {
				b.xRight = true
			} else if b.pos[0]+size[0] >= termSize[0] {
				b.xRight = false
			}
			if b.pos[1] <= 0 {
				b.yDown = true
			} else if b.pos[1]+size[1] >= termSize[1] {
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
		}

		// draw phase
		term.Reset()
		for i := range boxes {
			boxes[i].draw(term)
		}
		term.Flush()
		for i := range boxes {
			if i == 0 {
				continue
			}
			basicLine(term, boxes[i-1].box, boxes[i].box)
		}
		term.Flush()
	}
}
