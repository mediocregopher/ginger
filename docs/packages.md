# Packages

Ginger packages follow many of the same packaging rules as go packages. This
stems from ginger compiling down to go and needing to inter-op with go packages.

As discussed in the compilation doc, packages are defined as follows:

```
(. package "github.com/mediocregopher/awesome"

    (. def AwesomeThing "totally")

    (. defn AwesomeFunction []
        (: rand))
)
```

This expression does not have to appear in a particular folder heirarchy, and
multiple packages can appear in a single file. A package's definition can be
split up into multiple package statements across different files. This is
discussed more in the compilation doc.

## Variable/Function naming

Variables and functions follow the go rule of upper camel casing for public
variables/functions and lower cammel casing for private variables/functions.
This rule is enforced by the go compiler for both go packages and translated
ginger packages.

### Private

A private variable/function can only be referenced from within a package. If a
package is split into multiple parts across a project a private
variable/function defined in one part can be used in another part.

### Public

A public variable/function can be used within a package without any extra
embelishment (in the above example, `(: AwesomeFunction)` could simply be called
from within the package).

Outside of a package variables/functions can be used as follows:

```
(. package "show-and-tell"

    (. defn main []
        (: fmt.Println
            (: github.com/mediocregopher/awesome.AwesomeFunction)))
)
```

`show-and-tell.main` uses both the `Println` function from the `fmt` package and the
`AwesomeFunction` function from the `github.com/mediocregopher/awesome` package.
This syntax is rather cumbersome, however, and can be shortcutted using the
`alias` function in a package

```
(. package "show-and-tell"

    (. alias "github.com/mediocregopher/awesome" "aw")

    (. defn main []
        (: fmt.Println
            (: aw.AwesomeFunction)))
```

Like go, aliasing a package to `"."` imports it directly:

```
(. package "show-and-tell"

    (. alias "github.com/mediocregopher/awesome" ".")

    (. defn main []
        (: fmt.Println
            (: AwesomeFunction)))
```

Aliasing a package requires that you use it in the package you've aliased it in,
unless you alias to `"_"`.

## Idiomatic package usage

While it is not a requirement that your package namespaces follow the directory
heierarchy they show (in fact, you could have an entire project, with multiple
packages, all within a giant flat file), it's definitely recommended that you
do. It will make the code much easier to create a mental map of for newcomers to
it.

## Circular dependencies

Go enforces that a package may not have circular dependencies. That is,
`packageA` may not import `packageB` while `packageB` also imports `packageA`.
Ginger will also, be way of being translated to go, also enforce this rule.
