package expr

// MacroFn is a compiler function which takes in an existing Expr and returns
// the llvm Value for it
type MacroFn func(BuildCtx, Ctx, Expr) Expr

// Ctx contains all the Macros and Identifiers available. A Ctx also keeps a
// reference to the global context, which has a number of macros available for
// all contexts to use.
type Ctx struct {
	global *Ctx
	macros map[Macro]MacroFn
	idents map[Identifier]Expr
}

// NewCtx returns a blank context instance
func NewCtx() Ctx {
	return Ctx{
		global: globalCtx,
		macros: map[Macro]MacroFn{},
		idents: map[Identifier]Expr{},
	}
}

// Macro returns the MacroFn associated with the given identifier, or panics
// if the macro isn't found
func (c Ctx) Macro(m Macro) MacroFn {
	if fn := c.macros[m]; fn != nil {
		return fn
	}
	if fn := c.global.macros[m]; fn != nil {
		return fn
	}
	panicf("macro %q not found in context", m)
	return nil
}

// Identifier returns the llvm.Value for the Identifier, or panics
func (c Ctx) Identifier(i Identifier) Expr {
	if e := c.idents[i]; e != nil {
		return e
	}
	// The global context doesn't have any identifiers, so don't bother checking
	panicf("identifier %q not found", i)
	panic("go is dumb")
}

// Copy returns a deep copy of the Ctx
func (c Ctx) Copy() Ctx {
	cc := Ctx{
		global: c.global,
		macros: make(map[Macro]MacroFn, len(c.macros)),
		idents: make(map[Identifier]Expr, len(c.idents)),
	}
	for m, mfn := range c.macros {
		cc.macros[m] = mfn
	}
	for i, e := range c.idents {
		cc.idents[i] = e
	}
	return cc
}

// Bind returns a new Ctx which is a copy of this one, but with the given
// Identifier bound to the given Expr. Will panic if the Identifier is already
// bound
func (c Ctx) Bind(i Identifier, e Expr) {
	if _, ok := c.idents[i]; ok {
		panicf("identifier %q is already bound", i)
	}
	c.idents[i] = e
}
