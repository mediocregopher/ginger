# Functions

Functions are first-class citizens in ginger. The simplest anonymous function
can be created like so:

```
(. fn [x]
    (: + x 1))
```

This function takes in a single number and returns that number plus one. It
could be assigned to a variable on the package using `def`, or by using the
shortcut `defn`. The following two statements are equivalent:

```

(def incr
    (. fn [x]
        (: + x 1)))

(. defn incr [x]
    (: + x 1))
```

## Variadic function

Functions may take in any number of variables using the `...` syntax, similar to
how go does variadic functions. The variadic variable must be the last in the
function's argument list, and is used as a list inside of the function. A list
may also be used as the input for a variadic function, also using the `...`
syntax.

The following is an example of both defining a variadic function and both ways
of using it. `+` is a variadic function in this example:

```
(. defn avg [...x]
    (/
        (+ x...)
        (len x)))

(: fmt.Println (avg 1 2 3 4 5))
```
