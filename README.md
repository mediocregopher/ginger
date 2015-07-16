# Ginger

A lisp-like language built on the go programming language. The ideas are still a
work-in-progress, and this repo is where I'm jotting down my notes:

# Language manifesto

* Anything written in go should be writeable in ginger in as many lines or
  fewer.

* When deciding whether to be more go-like or more like an existing lisp
  language, err on being go-like.

* The fewer built-in functions, the better. The standard library should be
  easily discoverable and always importable so helper functions can be made
  available.

* When choosing between adding a syntax rule/datatype and not adding a feature,
  err on not adding the feature.

* It is not a goal to make ginger code be usable from go code.

* Naming should use words instead of symbols, except when those symbols are
  existing go operators.

* Overloading functions should be used as little as possible. Possibly not at
  all

# Documentation

See the [docs](/docs) folder for more details. Keep in mind that most of ginger
is still experimental and definitely not ready for the spotlight.

Here is a list of the docs more or less in the order they should be read for a
complete overview of the language:

* [syntax](/docs/syntax.md)
* [functions](/docs/functions.md)
* [compilation](/docs/compilation.md)
* [packages](/docs/packages.md)
* [go-interop](/docs/go-interop.md)
