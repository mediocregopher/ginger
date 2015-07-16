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
(. def incr
    (. fn [x]
        (: + x 1)))

(. defn incr [x]
    (: + x 1))
```

## Function returns

### Single returns

A function returns whatever the last expression in its execution returned. For
example:

```
(. defn abs [x]
    (. if (: >= x 0)
        x
        (: * -1 x)))
```

This function will return the argument `x` if it is greater than or equal to 0,
or `(: * -1 x)` if it's not.

### Multiple returns

A function which wishes to return multiple arguments should return them as a
vector or list of arguments. The `let` function, which can be used to define
temporary variables in a scope, can deconstruct these multiple-returns:

```
# Returns two numbers which sum up to 10
(. defn sum-10 []
    [4 6])

(. let [[foo bar] (: sum-10)
    (: fmt.Printf "%d + %d = 10\n" foo bar))
```

Functions defined within a go library which return multiple values can also be
deconstructed in a `let` statement:

```
(. let [[conn err] (: net.Dial "tcp" "localhost:80")]
    # do stuff)
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

## Tail-recursion

TODO
