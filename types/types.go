// This package describes ginger's base types and the interfaces covered by them
package types

// Elem is a generic type which can be used as a wrapper type for all ginger
// types, both base types and data structures
type Elem interface {
}

type String string

type Integer int

type Float float32

type Char rune

type Error error
