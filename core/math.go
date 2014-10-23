package core

import (
	"fmt"

	"github.com/mediocregopher/ginger/seq"
	"github.com/mediocregopher/ginger/types"
)

type mathReduceFn func(types.Elem, types.Elem) types.Elem

func mathReduce(fn mathReduceFn, zero types.Elem, s seq.Seq) types.Elem {
	reduceFn := func(acc, el types.Elem) (types.Elem, bool) {
		return fn(acc, el), false
	}
	return seq.Reduce(reduceFn, zero, s)
}

func plusInt(a, b types.Elem) types.Elem {
	return types.GoType{Int(a) + Int(b)}
}

func plusInt8(a, b types.Elem) types.Elem {
	return types.GoType{Int8(a) + Int8(b)}
}

func plusInt16(a, b types.Elem) types.Elem {
	return types.GoType{Int16(a) + Int16(b)}
}

func plusInt32(a, b types.Elem) types.Elem {
	return types.GoType{Int32(a) + Int32(b)}
}

func plusInt64(a, b types.Elem) types.Elem {
	return types.GoType{Int64(a) + Int64(b)}
}

func plusUint(a, b types.Elem) types.Elem {
	return types.GoType{Uint(a) + Uint(b)}
}

func plusUint8(a, b types.Elem) types.Elem {
	return types.GoType{Uint8(a) + Uint8(b)}
}

func plusUint16(a, b types.Elem) types.Elem {
	return types.GoType{Uint16(a) + Uint16(b)}
}

func plusUint32(a, b types.Elem) types.Elem {
	return types.GoType{Uint32(a) + Uint32(b)}
}

func plusUint64(a, b types.Elem) types.Elem {
	return types.GoType{Uint64(a) + Uint64(b)}
}

func plusFloat32(a, b types.Elem) types.Elem {
	return types.GoType{Float32(a) + Float32(b)}
}

func plusFloat64(a, b types.Elem) types.Elem {
	return types.GoType{Float64(a) + Float64(b)}
}

func plusComplex64(a, b types.Elem) types.Elem {
	return types.GoType{Complex64(a) + Complex64(b)}
}

func plusComplex128(a, b types.Elem) types.Elem {
	return types.GoType{Complex128(a) + Complex128(b)}
}

func plusString(a, b types.Elem) types.Elem {
	return types.GoType{String(a) + String(b)}
}

func Plus(s seq.Seq) types.Elem {
	if seq.Empty(s) {
		return types.GoType{0}
	}

	first, _, _ := s.FirstRest()
	var fn mathReduceFn
	var zero types.Elem
	switch first.(types.GoType).V.(type) {
	case int:
		fn = plusInt
		zero = types.GoType{int(0)}
	case int8:
		fn = plusInt8
		zero = types.GoType{int8(0)}
	case int16:
		fn = plusInt16
		zero = types.GoType{int16(0)}
	case int32:
		fn = plusInt32
		zero = types.GoType{int32(0)}
	case int64:
		fn = plusInt64
		zero = types.GoType{int64(0)}
	case uint:
		fn = plusInt
		zero = types.GoType{uint(0)}
	case uint8:
		fn = plusInt8
		zero = types.GoType{uint8(0)}
	case uint16:
		fn = plusInt16
		zero = types.GoType{uint16(0)}
	case uint32:
		fn = plusInt32
		zero = types.GoType{uint32(0)}
	case uint64:
		fn = plusInt64
		zero = types.GoType{uint64(0)}
	case float32:
		fn = plusFloat32
		zero = types.GoType{float32(0)}
	case float64:
		fn = plusFloat64
		zero = types.GoType{float64(0)}
	case complex64:
		fn = plusComplex64
		zero = types.GoType{complex64(0)}
	case complex128:
		fn = plusComplex128
		zero = types.GoType{complex128(0)}
	case string:
		fn = plusString
		zero = types.GoType{string(0)}
	default:
		panic(fmt.Sprintf("$#v cannot have plus called on it", first))
	}

	return mathReduce(fn, zero, s)
}
