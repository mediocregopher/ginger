package expr

import "llvm.org/llvm/bindings/go/llvm"

// MacroFn is a compiler function which takes in an existing Expr and returns
// the llvm Value for it
type MacroFn func(BuildCtx, Expr) Expr

// Ctx contains all the Macros and Identifiers available. A Ctx also keeps a
// reference to the global context, which has a number of macros available for
// all contexts to use.
type Ctx struct {
	global *Ctx
	macros map[Macro]MacroFn
	idents map[Identifier]llvm.Value
}

// NewCtx returns a blank context instance
func NewCtx() *Ctx {
	return &Ctx{
		global: globalCtx,
		macros: map[Macro]MacroFn{},
		idents: map[Identifier]llvm.Value{},
	}
}

// GetMacro returns the MacroFn associated with the given identifier, or panics
// if the macro isn't found
func (c *Ctx) GetMacro(m Macro) MacroFn {
	if fn := c.macros[m]; fn != nil {
		return fn
	}
	if fn := c.global.macros[m]; fn != nil {
		return fn
	}
	panicf("macro %q not found in context", m)
	return nil
}

// GetIdentifier returns the llvm.Value for the Identifier, or panics
func (c *Ctx) GetIdentifier(i Identifier) (llvm.Value, bool) {
	if v, ok := c.idents[i]; ok {
		return v, true
	}
	// The global context doesn't have any identifiers, so don't bother checking
	return llvm.Value{}, false
}

// NewWith returns a new Ctx instance which imports the given macros from the
// parent
//func (c *Ctx) NewWith(mm ...Macro) *Ctx {
//	nc := &Ctx{
//		global: c.global,
//		macros: map[Macro]MacroFn{},
//	}
//	for _, m := range mm {
//		fn := c.macros[m]
//		if fn == nil {
//			panicf("no macro %q found in context", m)
//		}
//		nc.macros[m] = fn
//	}
//	return nc
//}
