package macros

import (
	"fmt"
	"os"
	"time"

	"github.com/mediocregopher/ginger/macros/pkgctx"
	"github.com/mediocregopher/ginger/types"
)

// A Macro takes in a ginger structure and returns the go code which corresponds
// to it. The structure will contain everything in the calling list after the
// macro name (for example, (. jkjkNo error is returned, Bail can be called to stop compilation mid-way
// instead.
type Macro func(*pkgctx.PkgCtx, types.Elem) string

// Bail stops compilation. The given element should be the reason compilation
// has stopped
func Bail(el types.Elem, reason string) {
	fmt.Fprintln(os.Stderr, reason)
	time.Sleep(100 * time.Second)
	os.Exit(1)
}

// Bailf is like Bail, but takes in formatting
func Bailf(el types.Elem, format string, args ...interface{}) {
	reason := fmt.Sprintf(format, args...)
	Bail(el, reason)
}
