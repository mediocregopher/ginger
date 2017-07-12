# Ginger - holy fuck again?

## The final result. A language which can do X

- Support my OS
    - Compile on many architectures
    - Be low level and fast (effectively c-level)
    - Be well defined, using a simple syntax
    - Extensible based on which section of the OS I'm working on
    - Good error messages

- Support other programmers and other programming areas
    - Effectively means able to be used in most purposes
    - Able to be quickly learned
    - Able to be shared
        - Free
        - New or improved components shared between computers/platforms/people

- Support itself
    - Garner a team to work on the compiler
        - Team must not require my help for day-to-day
        - Team must stick to the manifesto, either through the design or through
          trust

## The language: A manifesto, defines the concept of the language

- Quips
    - Easier is not better

- Data as the language
    - Differentiation between "syntax" and "language", parser vs compiler
        - Syntax defines the form which is parsed
            - The parser reads the syntax forms into data structures
        - Language defines how the syntax is read into data structures and
          "understood" (i.e. and what is done with those structures).
            - A language maybe have multiple syntaxes, if they all parse into
              the same underlying data structures they can be understood in the
              same way.
            - A compiler turns the parsed language into machine code. An
              interpreter performs actions directly based off of the parsed
              language.

- Types, instances, and operations
    - A language has a set of elemental types, and composite types
        - "The type defines the [fundamental] operations that can be done on the
          data, the meaning of the data, and the way values of that type can be
          stored"
        - Elemental types are all forms of numbers, since numbers are all a
          computer really knows
        - Composite types take two forms:
            - Homogeneous: all composed values are the same type (arrays)
            - Heterogeneous: all composed values are different
                - If known size and known types per-index, tuples
                - A 0-tuple is kind of special, and simply indicates absence of
                  any value.
        - A third type, Any, indicates that the type is unknown at compile-time.
          Type information must be passed around with it at runtime.
        - An operation has an input and output. It does some action on the input
          to produce the output (presumably). An operation may be performed as
          many times as needed, given any value of the input type. The types of
          both the input and output are constant, and together they form the
          operation's type.
    - A value is an instance of a type, where the type is known at compile-time
      (though the type may be Any). Multiple values may be instances of the same
      type. E.g.: 1 and 2 are both instances of int
        - A value is immutable
        - TODO value is a weird word, since an instance of a type has both a
          type and value. I need to think about this more. Instance might be a
          better name

- Stack and scope
    - A function call operates within a scope. The scope had arguments passed
      into it.
    - When a function calls another, that other's scope is said to be "inside"
      the caller's scope.
    - A pure function only operates on the arguments passed into it.
    - A pointer allows for modification outside of the current scope, but only a
      pointer into an outer scope. A function which does this is "impure"

- Built-in
    - Elementals
        - ints (n-bit)
        - tuples
        - stack arrays
            - indexable
            - head/tail
            - reversible (?)
            - appendable
        - functions (?)
        - pointers (?)
        - Any (?)
    - Elementals must be enough to define the type of a variable
    - Ability to create and modify elmental types
        - immutable, pure functions
    - Other builtin functionality:
        - Load/call linked libraries
    - Comiletime macros
        - Red/Blue

- Questions
    - Strings need to be defined in terms of the built-in types, which would be
      an array of lists. But this means I'm married to that definition of a
      string, it'd be difficult for anyone to define their own and have it
      interop. Unless "int" was some kind of macro type that did some fancy
      shit, but that's kind of gross.
    - An idea of the "equality" of two variables being tied not just to their
      value but to the context in which they were created. Would aid in things
      like compiler tagging.
    - There's a "requirement loop" of things which need figuring out:
        - function structure
        - types
        - seq type
        - stack/scope
    - Most likely I'm going to need some kind of elemental type to indicate
      something should happen at compile-time and not runtime, or the other way
      around.

## The roadmap: A plan of action for tackling the language
