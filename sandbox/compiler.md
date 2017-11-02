# Types

## Axioms

The syntax described only applies within this document.

- All defined things below are values, all values have a type, all types have a
  definition which is itself a value.

- A type definition is displayed as a value wrapped in angle brackets, like
  `<int>`.

- A type definition with no value, `<>`, is the empty type

- A type definition may have more than one type, as in `<int,string>`, to
  indicate that a value with that type is actually a combination of each
  type in sequence (i.e. a tuple)

    - Tuples can be wrapped in parenthesis to indicate sub-groupings. I.e.
      `<string,(int,int)>` is a tuple of 2 sub-types, the first being a
      `<string>` and the second being a tuple of 2 `<ints>`.

- Any type definition can be used in place of `<any>`.

- `1` is an example of a value of type `<int>`.

- The type of value `<int>` is `<typedef,int>`.
- 
- Any `lowercaseAlphaNumeric` string is an atom value, of type `<atom>`.

- `V<someType>` is a placeholder for a value of type `<someType>`. It is used
  when matching a pattern.

- `(declareType, V<any>)` is understood to declare a type definition. A type
  definition can then be used as `<typedef>`.

- `(declareFunction, V<atom>, <any>, <any>)` is a tuple understood to declare a
  function, named after the atom, where the type definitions of the input and
  output are also given.

## Type declarations.

```
(declareType, atom)
(declareType, bool)
(declareType, int)

// general purpose functions for working with all types.
(declareFunction, concat, <any, any>,  <any>)
(declareFunction, slice, <any,int,int>, <any>)
(declareFunction, len, <any>, <int>)
(declareFunction, eq, <any,any>, <bool>)

// functions for working with integers
(declareFunction, plus, <int,int>, <int>)
(declareFunction, mult, <int,int>, <int>)
(declareFunction, minus, <int,int>, <int>)
// these two may return false if divide by zero
(declareFunction, div, <int,int>, <int,bool>)
(declareFunction, mod, <int,int>, <int,bool>)

// a general iterator
(declareType, (iter,any))
(declareFunction, next, <iter,any>, <(iter,any),any,bool>)

// TODO structurally, what's the difference between `<int,int>` and
// `<iter,any>`? the latter's first element isn't a valid typedef on its own,
// but other than that there seems to be no difference?

//(declareCompound, graph, T)
//(declareFunction, addEdge, (tup,(graph,T),T,T), (tup,graph,T))
//(declareFunction, rmEdge, (tup,(graph,T),T,T), (tup,graph,T))
//// the order of elements returned by parents/children is the same as the order
//// the edges between the nodes were added.
//(declareFunction, parents, (tup,(graph,T),T), (tup,(iter,T)))
//(declareFunction, children, (tup,(graph,T),T), (tup,(iter,T)))
//(declareFunction, has, (tup,(graph,T),T), bool)
//```
