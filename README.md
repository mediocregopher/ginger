# Ginger - I'll get it right this time

## A note on compile-time vs runtime

Ginger is a language whose primary purpose is to be able to describe and compile
itself. A consequence of this is that it's difficult to describe the actual
process by which compiling is done without first describing the built-in types,
but it's also hard to describe the built-in types without first describing the
process by which compiling is done. So I'm going to do one, then the other, and
I ask you to please bear with me.

## The primitive types

Ginger is a language which encompasses itself. That means amongst the "normal"
primitives a language is expected to have it also has a couple which are used
for macros (which other languages would not normally expose outside of the
compiler's implementation).

```
// These are numbers
0
1
2
3

// These are strings
"hello"
"world"
"how are you?"

// These are identifiers. Values at runtime are bound to
// identifiers, such that whenever an identifier is used in a non-macro
// statement that value will be replaced with it
foo
barBaz
biz_buz

// These are macro identifiers. They are like identifiers, except they start
// with percent signs, and they represent operations or values which only exist
// at compile-time There are a number of builtin macros, but they can also be
// user-defined. We'll see more of them later
%foo
%barBaz
%biz_buz
```

## The data structures

Like the primitives, ginger has few built-in data structures, and the ones it
does have are primarily used to implement itself.

```
// These are tuples. Each is a unique and different type, based on its number of
// elements, the type of each element, and the order those types are in. The
// type of a tuple must be known at compile-time
1, 2
4, "foo", 5

// These are arrays. Their elements must be of the same type, but their length
// can be dynamically determined at runtime. The type of an array is only
// determined by the type of its elements, which must be known at compile-time
[1, 2, 3]
["a", "b", "c"]

// These are statements. A statement is a pair of things, the first being a
// macro identifier, the second being an argument (usually a tuple).
%add 1,2
%incr 1
```

There is a final data structure, called a block, which I haven't come up with a
special sytax for yet, and will be discussed later.

## Parenthesis

A pair of parenthesis can be used to enclose any type for clarity. For example:

```
// No parenthesis
%add 1, 2

// Parenthesis around the argument (the tuple)
%add (1, 2)

// Parenthsis around the statement
(%add 1, 2)

// Parenthesis around everything
(%add (1, 2))
```

## Compilation

Ginger programs are taken as a list of statements (as in, the primitive types
we've defined already).

During compilation each statement is looked at, first its arguments then its
operator. The arguments are "resolved" first, so that they have only primitive
types that aren't macros, statements or blocks. Then that is passed into the
macro operator which may output a further statement, or may change something in
the context of compilation (e.g. the set of identifier bindings), or both. This
is done until the statement contains no more macros to run, at which point the
process repeats at the next statement.

### Example

It's difficult to see this without examples, imo. So here's some example code,
with explanatory comments:

```
// %bind is a macro, it takes in a tuple of an identifier and some value, and
// assigns that value to the identifier at runtime
%bind a, 1

// %add takes in a tuple of numbers or identifiers and will return the sum of
// them. Here we're then binding that sum (3) to the identifier b.
%bind b, (%add a, 2)

// The previous two example are fairly simple, but do something subtle. A ginger
// program starts as a list of statements, and must continue to be a list of
// statements after all macros are run. Each of the above is a macro statement
// which returns a "runtime statement", i.e. a construct representing something
// which will happen at runtime. But they are of type `statement` nonetheless,
// so running these macros does not change the overall type of the program (a
// list of statements)

// Creates an identifier c and returns it. This can't be included at this point,
// because it doesn't return a statement of any sort.
// %ident "c"

// This first creates an identifier a, which is then part of a tuple (a, 2).
// This tuple is used in a further tuple, now (%add, (a, 2)). Remember, %add is
// simply a macro identifier at this point, it's not actually "run" because it's
// part of a tuple, not a statement, and as such can be passed around like any
// other primitive type.
//
// Finally, the tuple (%add, (a, 2)) is passed into %stmt, which creates a new
// statement from a tuple of an operation and an argument. So the statement
// (%add a, 2) is returned. Since this statement still has a macro, %add, that
// is then called, and it finally returns a runtime statement which adds the
// value a is bound to to 2>
%stmt %add, (%ident "a", 2)

// This is exactly equivalent to the above statement, except that it skips some
// redundant macro processing. They can be used interchangeably in all cases and
// situations.
%add a, 2
```

## Blocks

Thus far we've only been able to create code linearly, without much way to do
code-reuse or scoping or anything like that.

Blocks fix this. A block is composed of three lists:

- A list of identifiers which will be "imported" from the parent block (the top
  level list of list of statements is itself a block, psych!).

- A list of statements

- A list of identifiers which will be "exported" from the block into the parent

There is not yet a special syntax for blocks, but there is a macro operator to
make them, much like the ones for statements and identifiers:

```
%bind a, 2

%do (%block [a], [
    %bind b, (%add a, 3)
], [b])

%println b // prints 5
```

In the above we create a block which imports the `a` identifier, and exports the
`b` identifier that it creates internally. Note that we have to use `%do`
in order to actually "run" the block, since `%block` merely returns the block
structure, which is not a statement.

This seems kind of like a pain, and not much like a function. But combined with
other macros blocks can be used to implement your own function dispatch, so you
can add in variadic, defaults, named parameters, as well as implement closures,
type methods, and so forth, as needed and in the style desired.

## Final note

Keep in mind: blocks, statements, etc... are themselves data structures, and
given appropriate built-in macros they can be manipulated like any other data
structure. These are merely the building blocks for all other language features
(hopefully).

