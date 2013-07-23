# Ginger

A scripted lisp language with simple syntax, immutable data structures, concurrency built-in, and
minimal time between starting the runtime and actual execution.

# Documentation

Documentation is broken up into different parts, more or less in the order a newcomer should read them to get familiar
with the language.

* [syntax/data structures](/doc/syntax.md) - Ginger is a lisp-based language, so if you know the syntax
                                             for the available data structures you know the syntax for the language
                                             itself. There's very few data structures, meaning there's minimal syntax.
* [runtime](/doc/runtime.md)               - How to structure your ginger data to be runnable by the ginger interpreter.
                                             Includes execution, variable definition, scope, and function definition.
* [pattern matching](/doc/pattern.md)      - Deconstruction of data structures and case statements on them.
