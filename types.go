package ginger

type Expr struct {
	// [0-9]+
	Int int

	// true | false
	Bool bool

	// [Expr [, Expr]]
	Tuple []Expr

	// { [Statement (;\s)]* }
	Block []Expr

	// [Expr | Expr]
	Pipeline []Expr

	// [a-z]+
	Identifier string

	// Expr > Expr
	Statement *struct {
		Input Expr
		Into  Expr
	}
}
