# Ginger

A lisp-like language built on the go programming language. The ideas are still a
work-in-progress, and this repo is where I'm jotting down my notes.

# Goals

I have some immediate goals I'm trying to achieve with this syntax:

* Everything is strings (except numbers, functions, and data structures). There
  is no symbol type, atom type, keyword type, etc... they're all just strings.

* There is no `defmacro`. Macro creation and usage is simply an inherent feature
  of the language syntax.

# Walkthrough

This is a number which evalutates to 5:

```
5
```

This is a string, it can contain anything:

```
"! I'm the king of the world !"
```

This is a list. It evaluates to a linked-list of four strings:

```
("a" "b" "c" "d")
```

This is a vector of those same elements. It's like a list, but has some slightly
different properties. We'll mostly be using lists:

```
["a" "b" "c" "d"]
```

This is a string

```
"+"
```

`:` is the evaluator. A string beginning with `:` is evaluated to whatever it
references. This evaluates to a function which adds its arguments:

```
":+"
```

This evaluates to list whose elements are a function and two numbers:

```
(":+" 1 2)
```

A list prefixed with a `:` calls the first element as a function with the rest
of the elements as arguments. This evaluates to the number 5:

```
:(":+" 1 2)
```

A bare string (lacking in `"`) is a shortcut for that string prefixed by a `:`.
This is equivalent to the above:

```
:(+ 1 2)
```

The `fn` function can be used to define a new function. This evaluates to an
anonymous function which adds one to its argument and returns it:

```
:(fn [x]
    :(+ x 1))
```

The `def` function can be used to bind some value to a new variable. Note the
`.` instead of `:`. We'll cover that in a bit. This defines a variable `foo`
which evaluates to the string `"bar"`:

```
.(def foo "bar")
```

This defines a variable `incr` which evaluates to a function which adds one to
its argument:

```
.(def incr
    :(fn [x]
        :(+ x 1)))
```

This uses `defn` as a shortcut for the above:
```
.(defn incr [x]
    :(+ x 1))
```

There are also maps. A map's keys can be any value(?). A map's values can be any
value. This evaluates to a map with 2 key/val pairs:

```
{ "foo" foo
  "bar" :(incr 4) }
```

`.` is the half-evaluator. It only works on lists, and runs the function given
in the first argument with the unevaluated arguments (even if they have `:` in
front). You can generate new code to run on the fly (macros) using the normal
`fn`. This evaluates to a `let`-like function, except it forces you to use the
capitalized variable names in the body (utterly useless):

```
.(defn caplet [mapping body...]
    # elem-map maps over every element in a list, embedded or otherwise
    ::(elem-map
        :(fn [x]
            :(if (mapping :(slice x 1))
                (capitalize x)
                x))
        body))

#Usage
.(caplet [foo "this is foo"
          dog "this is dog"]
    :(println Foo)
    :(println Dog))
```
