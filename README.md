# Ginger

A lisp-like language built on the go programming language. The ideas are still a
work-in-progress, and this repo is where I'm jotting down my notes.

# Walkthrough

This is a number which evalutates to 5:

```
5
```

This is a string, as it contains no whitespace:

```
ImJustAString
```

This is also a string, it can contain anything:

```
"! I'm the king of the world !"
```

This is a list. It evaluates to a linked-list of its elements (not lisp-like!):

```
(a b c d)
```

This is a string:

```
+
```

This is a string being evaluated, it will return a function:

```
:+
```

This is a list, with a string and two numbers:

```
(+ 1 2)
```

This is a list, with a function and two numbers:

```
(:+ 1 2)
```

This is a list being evalutated, the first item in the list must be a function:

```
:(:+ 1 2)
```

This is a list in a list:

```
((a b) foo bar)
```

This is an anonymous function, it takes in the arguments `a` and `b`, and
returns their sum:

```
#((a b)
    :(:+ :a :b))
```

`map` takes in a function and a list, and calls the function on each item in the
list. The following will return `(1 2 3)`:

```
:(:map               ; This is a comment, it's ignored
    #((a) (:+ :a 1)) ; <- Increment function
    (0 1 2))
```

`name` names a value. After the following call, `:ted` will always evalutate to
`5`:

```
:(:name ted 5)
```

You can name any value, including functions:

```
:(:name increment #((a) :(:+ a 1)))
```

Now the `map` example above can be simplified down to:

```
:(:map :increment (0 1 2))
```
