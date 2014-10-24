# mathgen, or why I'm so so sorry

Go's type system is a double-edged sword, and the inner edge will definitely get
you when you go to do generic math. I opted against using the `reflect` library,
and instead went for codegen. So while the result is barely readable, it will at
least be somewhat faster, theoretically.

This module generates the `math.go` file which resides in the core library. The
`mathgen.tpl` file is where most of the magic happens. It is virtually
unreadable. The biggest reason for this is that I want it to be able to pass
through `go fmt` unphased, so that it doesn't end up getting changed back and
forth throughout the code's history.

## Running

You can run code gen by running `make gen` in the `core/` directory. The
`Makefile` in the root of the project should also reference this, so running
that in the root should also result in the file getting regenerated.
