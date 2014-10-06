# Go interop

Ginger translates down to go code, and many of its conventions and rules follow
from go's conventions and rules. In most cases these decisions were made to help
with interoperability with existing go code.

## Referencing go package variables/functions

See the package doc for more on this

## Types

Go types and ginger types share a lot of overlap:

* Ginger strings are of go's `string` type

* Ginger integers are of go's `int` type

* Ginger floats are of go's `float32` type

* Ginger characters are of go's `rune` type

* Ginger errors are of go's `error` type

## Casting

Each go type has a corresponding ginger casting function:

```
(: int64 5)
(: float64 5.5)
(: rune 'a')
```

## go-drop

the `go-drop` form can be used for furthur interoperability. The rationale
behind `go-drop` is that there are simply too many cases to be able to create
enough individual functions, or a few generic functions, that would cover all
cases. Instead we use a single function, `go-drop`, which lets us drop down into
go code and interact with it directly. There are a number of pre-made functions
which implement commonly needed behaviors, such as `StringSlice` and
`ByteSlice`, which cast from either go or ginger types into `[]string` and
`[]byte`, respectively.

```
(. go-drop
    "func StringSlice(v ginger.Elem) []string {
        ret := []string{}
        // do some stuff
        return ret
    }")
```
