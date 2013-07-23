# Pattern Matching

Pattern matching in ginger is used both to assign variables and test for the contents of them.

## Assignment

The `=` function's job is to assign values to variables. In the simplest case:
```
(= Foo 5)
```

Assigns the value `5` to the variable named `Foo`. We can assign more complicated values to variables
as well:
```
(= Foo [1 2 3 4])
```

### Deconstruction

With pattern matching, we can deconstruct the value on the right side and assign its individual parts
to their own individual variables:
```
[
    (= [Foo Bar [Baz Box] Bus] [1 a [3 4] [5 6]])
    (println Foo)
    (println Bar)
    (println Baz)
    (println Box)
    (println Bus)
]
```

The above would print out:
```
1
a
3
4
[5 6]
```

We can also deconstruct a previously defined variable:
```
[
    (= Foo [a b c])
    (= [A B C] Foo)
]
```

The above would assign `a` to `A`, `b` to `B`, and `c` to `C`.

### No Match

What happens if we try to pattern match a value that doesn't match the left side?:
```
(= [Foo Bar Baz Box] [1 2 3]) ; => [error nomatch]
```

The value itself is a vector with three items, but we want a vector with four! This returns an error
and none of the variables were assigned.

### Blanks

If there's a variable we expect the right side to have, but we don't actually want to do anything with
it, we can use `_`:
```
(= [Foo _ Baz] [1 2 3])
```

The above would assign `1` to `Foo` and `2` to `Baz`.

### Left-side checks

The left side isn't exclusively used for assignment, it can also be used to check certain values on
the right:
```
(= [1 Foo 2 Bar] [1 a 2 b]) ; => success
(= [1 Foo 2 Bar] [1 a 3 b]) ; => [error nomatch]
```

If a variable is already defined it can be used as a check as well:
```
[
    (= Foo 1)
    (= [Foo Foo Bar] [1 1 2]) ; => success
    (= [Foo Foo Baz] [1 2 3]) ; => [error nomatch]
]
```

Finally, undefined variables on the left side that are used more than once are ensured that all
instances have the same value:
```
[
    (= [Foo Foo Foo] [foo foo foo]) ; => success
    (= [Bar Bar Bar] [foo bar bar]) ; => [error nomatch]
]
```

### Vectors

What if we don't know the full length of the vector, but we only want to retrieve the first two values
from it?:
```
(= [Foo Bar -] [1 2 3 4 5])
```

The above would assign `1` to `Foo` and `2` to `Bar`. `-` is treated to mean "zero or more items in the sequence
that we don't care about".

What if you want to match on the tail end of the vector?:
```
(= [- Bar Foo] [1 2 3 4 5])
```

The above would assign `4` to Bar and `5` to `Foo`.

Can we do both?:
```
(= [Foo - Bar] [1 2 3 4 5])
```

The above would assign `1` to `Foo` and `5` to `Bar`.

### Lists

Everything that applies to vectors applies to lists, as far as pattern matching goes:
```
(= (Foo Bar Baz -) (l (foo bar baz)))
```

In the above `foo` is assigned to `Foo`, `bar` to `Bar`, and `baz` to `Baz`.

Note that the right side needs the `l` literal wrapper to prevent evaluation, while
the left side does not. Nothing on the left will ever be evaluated, if you want it
to be dynamically defined you'll have to use a macro.

### Maps

Pattern matching can be performed on maps as well:
```
(= { Foo Bar } { a b }) ; => success
(= { Foo Bar } { a b c d }) ; => [error nomatch]
(= { Foo Bar - - } { a b c d }) ; => success
```

In the last case, the important thing is that the key is `-`. This tells the matcher
that you don't care about any other keys that may or may not exist. The value can be
anything, but `-` seems semantically more correct imo.

## Case

We can deconstruct and assign variables using `=`, and even test for their contents
to some degree since `=` returns either `success` or `[error nomatch]`. However
that's not very convenient. For this we have the `case` statement:
```
[
    (= Foo bird)
    (case Foo
        bird  (println "It's a bird!")
        plane (println "It's a plane!")
        _     (println "I don't know what it is!"))
]
```

In the above the output would be:
```
It's a bird!
```

`case` returns the result of whatever it matches:
```
[
    (= Animal bird)
    (case Animal
        bird chirp
        dog  woof
        lion roar) ; => chirp
]
```

`case` runs through all the different patterns in order, attempting to match on one
using an `=` on the backend. If none can be found it returns `[error nomatch]`. Since
`_` matches everything it can be used for any default statement.

Deconstruction works with `case` too:
```
(case [a b c]
    [Foo c d] [foo Foo]
    [Bar b c] [bar Bar]
    _         baz) ; => [bar a]
```

### Guards

What's a case statement without guards? Nothing, that's what! Instead of evaluating the statement
following the pattern, if that statement is `when` followed by a statement returning a boolean,
finally followed by the actual statement, then boolean test must evaluate to `true` or the
case is skipped. The boolean test can use variables from the pattern that was matched:
```
(case [five 5]
    (Name Num) when (> Num 2) (println Name "is greater than two")
    (Name Num)                (println Name "is less than or equal to two")
    _                         (println "wut?"))
```
