package lang

// Commonly used Terms
var (
	// Language structure types
	AAtom  = Atom("atom")
	AConst = Atom("const")
	ATuple = Atom("tup")
	AList  = Atom("list")

	// Match shortcuts
	AUnder    = Atom("_")
	TDblUnder = Tuple{AUnder, AUnder}

	// VM commands
	AInt = Atom("int")
	AAdd = Atom("add")
)
