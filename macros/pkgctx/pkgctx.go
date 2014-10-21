package pkgctx

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
