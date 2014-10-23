package core

import (
	"github.com/mediocregopher/ginger/types"
)

func Int(e types.Elem) int {
	return e.(types.GoType).V.(int)
}

func Int8(e types.Elem) int8 {
	return e.(types.GoType).V.(int8)
}

func Int16(e types.Elem) int16 {
	return e.(types.GoType).V.(int16)
}

func Int32(e types.Elem) int32 {
	return e.(types.GoType).V.(int32)
}

func Int64(e types.Elem) int64 {
	return e.(types.GoType).V.(int64)
}

func Uint(e types.Elem) uint {
	return e.(types.GoType).V.(uint)
}

func Uint8(e types.Elem) uint8 {
	return e.(types.GoType).V.(uint8)
}

func Uint16(e types.Elem) uint16 {
	return e.(types.GoType).V.(uint16)
}

func Uint32(e types.Elem) uint32 {
	return e.(types.GoType).V.(uint32)
}

func Uint64(e types.Elem) uint64 {
	return e.(types.GoType).V.(uint64)
}

func Float32(e types.Elem) float32 {
	return e.(types.GoType).V.(float32)
}

func Float64(e types.Elem) float64 {
	return e.(types.GoType).V.(float64)
}

func Complex64(e types.Elem) complex64 {
	return e.(types.GoType).V.(complex64)
}

func Complex128(e types.Elem) complex128 {
	return e.(types.GoType).V.(complex128)
}

func Bool(e types.Elem) bool {
	return e.(types.GoType).V.(bool)
}

func Byte(e types.Elem) byte {
	return e.(types.GoType).V.(byte)
}

func Rune(e types.Elem) rune {
	return e.(types.GoType).V.(rune)
}

func String(e types.Elem) string {
	return e.(types.GoType).V.(string)
}

func Error(e types.Elem) error {
	return e.(types.GoType).V.(error)
}
