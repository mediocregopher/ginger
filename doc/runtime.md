# Runtime

Any ginger data-structure can be put into a ginger file. The data itself has no executional meaning on its own, but if
the data is properly formed it can be parsed by the ginger interpreter and a thread of execution can be started. For
example, if I put the following in a file called `test.gng`:
```
(1 2 3)
```

It doesn't have any meaning, it's just a list. However, if you put the following:
```
(add 1 2 3)
```

and run `ginger test.gng` then a program can be interpreted from the given data. This file describes how given data can
be formed into a valid program.

## Eval

Ginger evaluation is done the same was as other lisp languages: the first item in a list is the function name, the rest
of the items in the list are arguments to the function. In the example above, `add` is the function name, and `1`, `2`,
and `3` are arguments to the function.

Arguments to a function can be functions to be eval'd themselves. An equivalent to the example above would have been:
```
(add 1 2 (add 1 2))
```

## Doing multiple things

It's not very useful to only be able to do one thing. A vector of lists is interpreted into sequentially eval'ing each
item in the vector. For (a trivial) example:
```
[ (add 1 2)
  (sub 4 (add 1 2))
  (mul 8 0) ]
```

## Variables/Scope

The above example does a few things, but it repeats itself in the second part (with the `sub`). If we could save the
result of the addition to a variable that would be awesomesauce. The `=` function does this! It makes it so that the
first argument is equivalent to the evaluation of the second argument. Variables' can not be re-defined to be another
value, and their first letters must be upper-case, thisis how they're differentiated from raw string literals.The above
could be re-written as:
```
[ (= AdditionResult (add 1 2))
  (sub 4 AdditionResult)
  (mul 8 0) ]
```

In the above example `AdditionResult` is a valid variable inside of the vector that contains its declaration, and any
vectors contained within that vector. For example:
```
[ (= AdditionResult (add 1 2))
  [ (sub 4 AdditionResult) ;This works
    (mul 8 0) ] ]

[ (add 4 AdditionResult) ] ;This does not work!
```

## Literals

We've determined that a list is interpreted as a function/arguments set, and an upper-case string is interpreted as a
variable name which will be de-referenced. What if we want to actually use these structures without eval'ing them? For
these cases we have the literal function `l`:
```
[ (concat (l (1 2 3)) (l (4 5 6))) ; => (1 2 3 4 5 6)
  (println (l "I start with a capital letter and I DONT CARE!!!")) ]
```

## Functions

### Anonymous

Anonymous functions are declared using the `fn` function. The first argument is a vector of argument names (remember,
all upper-case!) and the second is a list/vector to be eval'd. Examples:
```
[ (= Add3 (fn [Num] (add Num 3)))
  (Add3 4) ; => 7

  (= Add3Sub1
    (fn [Num] [
      (= Added (add Num 3))
      (sub Added 1)
    ]))
  (Add3Sub1 4) ] ; => 6
```

`fn` returns a function, which can be passed around and assigned like any other value. In the above examples the
functions are assigned to the `Add3` and `Add3Sub1` variables. Functions can also be passed into other functions as
arguments:
```
;DoTwice takes a function Fun and a number Num. it will call Fun twice on Num and return the result.
[ (= DoTwice
    (fn [Fun Num]
      (Fun (Fun Num))))

  (DoTwice (fn [Num] (add Num 1)) 3) ] ; => 5
```

### Defined

Defined functions attach the function definition to a string literal (lower-case) in the current scope. They are useful
as they support more features then an anonymous function, such as inline documentation. These extra features will be
documented elsewhere. To create a defined function use the `dfn` function:
```
[ (dfn add-four-sub-three [Num] [
    (= A (add Num 4))
    (sub A 3) ])
  (add-four-sub-three 4) ] ; => 5
```

## Namespaces

Namespaces give names to defined scopes which can be referenced from elsewhere. The best way to show this is with an
example:
```
[
(ns circle [
    (= Pi 3.14)
    (dfn area [R] (mul R R Pi)) ])

(circle/area 5) ; => 78.5
]
```

### Embedded namespaces

Namespaces can be embedded into a tree-like structure:
```
[
(ns math [
    (= Pi 3.14)
    (ns circle
        (dfn area [R] (mul R R Pi)))
    (ns square
        (dfn area [R] (mul R R)))
])

(math.circle/area 5) ; => 78.5
(math.square/area 5) ; => 25
]
```

In the above example `circle` and `square` are both sub-namespaces of the `math` namespace. In `circle` the variable
`Pi` is referenced. Ginger will look in the current scope for that variable, and when it's not found travel up the scope
tree. `Pi` is defined in `math`'s scope, so that value is used.

### Namespace resolution

Let's do a more complicated example to show how namespace resolution works:
```
[
(ns tlns [
    (ns alpha
        (= A "The first letter in the alphabet"))
    (ns facts
        (= BestLetter alpha/A))
])

(println tlns.facts/BestLetter)
]
```

In the above example `BestLetter` is defined inside `facts` to be the variable `A` which exists in the namespace `alpha`.
To resolve this variable ginger first looks inside `facts`'s scope for a namespace called `alpha`, doesn't find it, then
moves up to `tlns`'s scope, which does contain a namespace called `alpha`. Similaraly, to resolve `tlns.facts/BestLetter`
ginger first looks in that statements current scope for a namespace called `tlns`, which it finds, and looks in there
for `facts`, etc...

### Variable namespaces

A namespace is nothing more than a string literal. If a variable is used instead ginger will resolve the variable before
trying to resolve the namespace.
```
[
(ns tlns [
    (ns alpha
        (= A "The first letter in the alphabet"))])

(= NS1 tlns)
(= NS2 alpha)
(= NS NS1.NS2)
(println NS/A) ; This would print the message defined above
]
```
