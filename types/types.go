// This package describes ginger's base types and the interfaces covered by them
package types

// Elem is a generic type which can be used as a wrapper type for all ginger
// types, both base types and data structures
type Elem interface {
}

// Number can either be either an Int or a Float
type Number interface {
}

type Str string

type Int int

type Float float32

type Char rune

type Err error
