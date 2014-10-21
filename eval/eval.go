// The eval package encompasses all that is necesary to do runtime evaluation of
// ginger structures. These are different than macros in that they aren't turned
// into go code, instead the compiled go code evaluates them at runtime
package eval

import (
	"fmt"
	"os"
	"time"

	"github.com/mediocregopher/ginger/macros/pkgctx"
	"github.com/mediocregopher/ginger/seq"
	"github.com/mediocregopher/ginger/types"
)

// Evaler is a function which can be used inside of eval. It must take in its
// arguments as a sequence of Elems, and return a resulting Elem
type Evaler func(seq.Seq) types.Elem

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

var colon = types.GoType{":"}

// Eval takes in the pkgctx it is being executed in, as well as a single Elem to
// be evaluated, and returns the Elem it evaluates to
func Eval(p *pkgctx.PkgCtx, el types.Elem) types.Elem {
	l, ok := el.(*seq.List)
	if !ok {
		return el
	}

	first, rest, ok := l.FirstRest()
	if !ok || !first.Equal(colon) {
		return el
	}
	
	fnEl, args, ok := rest.FirstRest()
	if !ok {
		Bail(el, "Empty list after colon, no function given")
	}
	

	var fnName string
	if gt, ok := fnEl.(types.GoType); ok {
		fnName, _ = gt.V.(string)
	}

	if fnName == "" || fnName[0] != ':' {
		Bail(el, "Must give a function reference to execute")
	}
	fnName = fnName[1:]

	fn, ok := p.CallMap[fnName]
	if !ok {
		Bailf(el, "Unknown function name %q", fnName)
	}

	evalArgFn := func(el types.Elem) types.Elem {
		return Eval(p, el)
	}

	evaldArgs := seq.Map(evalArgFn, args)

	return fn.(Evaler)(evaldArgs)
}
