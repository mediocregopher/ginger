package expr

import "llvm.org/llvm/bindings/go/llvm"

type MacroFn func(*Ctx, LLVMCtx, Expr) (llvm.Value, bool)

// Ctx contains all the Macros and Identifiers available. A Ctx is based on the
// parent it was created from. If the current Ctx doesn't have a particular key
// being looked up, the parent is called instead, and so on. A consequence of
// this is that keys in the children take precedence over the parent's
type Ctx struct {
	Parent *Ctx
	Macros map[Macro]MacroFn

	LastVal llvm.Value
}

// GetMacro returns the first instance of the given of the given Macro found. If
// not found nil is returned.
func (c *Ctx) GetMacro(m Macro) MacroFn {
	if c.Macros != nil {
		if fn, ok := c.Macros[m]; ok {
			return fn
		}
	}
	if c.Parent != nil {
		return c.Parent.GetMacro(m)
	}
	return nil
}
