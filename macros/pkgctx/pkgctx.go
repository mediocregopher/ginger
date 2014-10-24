package pkgctx

import (
	"strings"

	"github.com/mediocregopher/ginger/seq"
	"github.com/mediocregopher/ginger/types"
)

const (
	coreAbs = "github.com/mediocregopher/ginger/core"
	coreAlias = "gingercore"
)

// PkgCtx is given to all macros and represents the package that they are being
// evaluated within.
type PkgCtx struct {

	// Packages describes the external packages imported by this one. Each key
	// is the absolute package path, the value is the alias for it (or empty
	// string for no alias)
	Packages map[string]string

	// CallMap is a map used by Eval for making actual calls dynamically. The
	// key is the string representation of the call to be used (for example,
	// "fmt.Println") and must agree with the aliases being used in Packages.
	// The value need not be set during actual compilation, but it is useful to
	// use it during testing
	CallMap map[string]interface{}
}

// Returns an empty PkgCtx
func New() *PkgCtx {
	return &PkgCtx{
		Packages: map[string]string{},
		CallMap:  map[string]interface{}{},
	}
}

// A returns a new PkgCtx, which is a copy of a merge between p and p2. p2's
// keys overwrite any conflicting keys in p. p and p2 are unaffected by this
// operation
func (p *PkgCtx) MergeLeft(p2 *PkgCtx) *PkgCtx {
	p3 := New()
	for pkg := range p.Packages {
		p3.Packages[pkg] = p.Packages[pkg]
	}
	for pkg := range p2.Packages {
		p3.Packages[pkg] = p2.Packages[pkg]
	}
	for fn := range p.CallMap {
		p3.CallMap[fn] = p.CallMap[fn]
	}
	for fn := range p2.CallMap {
		p3.CallMap[fn] = p2.CallMap[fn]
	}
	return p3
}

// Returns a copy of p
func (p *PkgCtx) Copy() *PkgCtx {
	return p.MergeLeft(New())
}

func (p *PkgCtx) PopulateFromCode(el types.Elem) bool {
	if s, ok := el.(seq.Seq); ok {
		return seq.Traverse(p.PopulateFromCode, s)
	}

	gt, ok := el.(types.GoType)
	if !ok {
		return true
	}

	str, ok := gt.V.(string)
	if !ok {
		return true
	}

	if len(str) < 2 {
		return true
	}

	if str[0] != ':' {
		return true
	}
	str = str[1:]

	// At this point str is a reference to something. We check if it's already
	// pointing somewhere first
	if _, ok = p.CallMap[str]; ok {
		return true
	}

	// If there isn't a '.' in the string, it's not directly referencing another
	// package. Since it isn't in this context either, it must be referencing
	// something in core
	var i int
	if i = strings.IndexRune(str, '.'); i < 1 {
		p.Packages[coreAbs] = coreAlias
		p.CallMap[str] = nil

		// TODO there needs to be a CodeGen interface or something. The
		// CallMap's value type needs to be that, because in all likelyhood the
		// CallMap will be directly translated to a map variable in the
		// generated package, with the values being different depending on what
		// they're being used for (external functions will be references to
		// those functions, local variables may just end up being just GoType's
		// of the actual value.
	}

	return true
}
