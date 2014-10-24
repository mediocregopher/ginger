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
	var first, zero types.Elem

	if seq.Empty(s) {
		return types.GoType{0}
	}
	first, _, _ = s.FirstRest()

	var fn mathReduceFn
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
		fn = plusUint
		zero = types.GoType{uint(0)}

	case uint8:
		fn = plusUint8
		zero = types.GoType{uint8(0)}

	case uint16:
		fn = plusUint16
		zero = types.GoType{uint16(0)}

	case uint32:
		fn = plusUint32
		zero = types.GoType{uint32(0)}

	case uint64:
		fn = plusUint64
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
		zero = types.GoType{string("")}

	default:
		panic(fmt.Sprintf("$#v cannot have Plus called on it", first))
	}

	return mathReduce(fn, zero, s)
}
func minusInt(a, b types.Elem) types.Elem {
	return types.GoType{Int(a) - Int(b)}
}

func minusInt8(a, b types.Elem) types.Elem {
	return types.GoType{Int8(a) - Int8(b)}
}

func minusInt16(a, b types.Elem) types.Elem {
	return types.GoType{Int16(a) - Int16(b)}
}

func minusInt32(a, b types.Elem) types.Elem {
	return types.GoType{Int32(a) - Int32(b)}
}

func minusInt64(a, b types.Elem) types.Elem {
	return types.GoType{Int64(a) - Int64(b)}
}

func minusUint(a, b types.Elem) types.Elem {
	return types.GoType{Uint(a) - Uint(b)}
}

func minusUint8(a, b types.Elem) types.Elem {
	return types.GoType{Uint8(a) - Uint8(b)}
}

func minusUint16(a, b types.Elem) types.Elem {
	return types.GoType{Uint16(a) - Uint16(b)}
}

func minusUint32(a, b types.Elem) types.Elem {
	return types.GoType{Uint32(a) - Uint32(b)}
}

func minusUint64(a, b types.Elem) types.Elem {
	return types.GoType{Uint64(a) - Uint64(b)}
}

func minusFloat32(a, b types.Elem) types.Elem {
	return types.GoType{Float32(a) - Float32(b)}
}

func minusFloat64(a, b types.Elem) types.Elem {
	return types.GoType{Float64(a) - Float64(b)}
}

func minusComplex64(a, b types.Elem) types.Elem {
	return types.GoType{Complex64(a) - Complex64(b)}
}

func minusComplex128(a, b types.Elem) types.Elem {
	return types.GoType{Complex128(a) - Complex128(b)}
}

func Minus(s seq.Seq) types.Elem {
	var first, zero types.Elem

	if seq.Empty(s) {
		panic("Minus cannot be called with no arguments")
	}
	zero, s, _ = s.FirstRest()
	first = zero

	var fn mathReduceFn
	switch first.(types.GoType).V.(type) {

	case int:
		fn = minusInt

	case int8:
		fn = minusInt8

	case int16:
		fn = minusInt16

	case int32:
		fn = minusInt32

	case int64:
		fn = minusInt64

	case uint:
		fn = minusUint

	case uint8:
		fn = minusUint8

	case uint16:
		fn = minusUint16

	case uint32:
		fn = minusUint32

	case uint64:
		fn = minusUint64

	case float32:
		fn = minusFloat32

	case float64:
		fn = minusFloat64

	case complex64:
		fn = minusComplex64

	case complex128:
		fn = minusComplex128

	default:
		panic(fmt.Sprintf("$#v cannot have Minus called on it", first))
	}

	return mathReduce(fn, zero, s)
}
func multInt(a, b types.Elem) types.Elem {
	return types.GoType{Int(a) * Int(b)}
}

func multInt8(a, b types.Elem) types.Elem {
	return types.GoType{Int8(a) * Int8(b)}
}

func multInt16(a, b types.Elem) types.Elem {
	return types.GoType{Int16(a) * Int16(b)}
}

func multInt32(a, b types.Elem) types.Elem {
	return types.GoType{Int32(a) * Int32(b)}
}

func multInt64(a, b types.Elem) types.Elem {
	return types.GoType{Int64(a) * Int64(b)}
}

func multUint(a, b types.Elem) types.Elem {
	return types.GoType{Uint(a) * Uint(b)}
}

func multUint8(a, b types.Elem) types.Elem {
	return types.GoType{Uint8(a) * Uint8(b)}
}

func multUint16(a, b types.Elem) types.Elem {
	return types.GoType{Uint16(a) * Uint16(b)}
}

func multUint32(a, b types.Elem) types.Elem {
	return types.GoType{Uint32(a) * Uint32(b)}
}

func multUint64(a, b types.Elem) types.Elem {
	return types.GoType{Uint64(a) * Uint64(b)}
}

func multFloat32(a, b types.Elem) types.Elem {
	return types.GoType{Float32(a) * Float32(b)}
}

func multFloat64(a, b types.Elem) types.Elem {
	return types.GoType{Float64(a) * Float64(b)}
}

func multComplex64(a, b types.Elem) types.Elem {
	return types.GoType{Complex64(a) * Complex64(b)}
}

func multComplex128(a, b types.Elem) types.Elem {
	return types.GoType{Complex128(a) * Complex128(b)}
}

func Mult(s seq.Seq) types.Elem {
	var first, zero types.Elem

	if seq.Empty(s) {
		return types.GoType{1}
	}
	first, _, _ = s.FirstRest()

	var fn mathReduceFn
	switch first.(types.GoType).V.(type) {

	case int:
		fn = multInt
		zero = types.GoType{int(1)}

	case int8:
		fn = multInt8
		zero = types.GoType{int8(1)}

	case int16:
		fn = multInt16
		zero = types.GoType{int16(1)}

	case int32:
		fn = multInt32
		zero = types.GoType{int32(1)}

	case int64:
		fn = multInt64
		zero = types.GoType{int64(1)}

	case uint:
		fn = multUint
		zero = types.GoType{uint(1)}

	case uint8:
		fn = multUint8
		zero = types.GoType{uint8(1)}

	case uint16:
		fn = multUint16
		zero = types.GoType{uint16(1)}

	case uint32:
		fn = multUint32
		zero = types.GoType{uint32(1)}

	case uint64:
		fn = multUint64
		zero = types.GoType{uint64(1)}

	case float32:
		fn = multFloat32
		zero = types.GoType{float32(1)}

	case float64:
		fn = multFloat64
		zero = types.GoType{float64(1)}

	case complex64:
		fn = multComplex64
		zero = types.GoType{complex64(1)}

	case complex128:
		fn = multComplex128
		zero = types.GoType{complex128(1)}

	default:
		panic(fmt.Sprintf("$#v cannot have Mult called on it", first))
	}

	return mathReduce(fn, zero, s)
}
func divInt(a, b types.Elem) types.Elem {
	return types.GoType{Int(a) / Int(b)}
}

func divInt8(a, b types.Elem) types.Elem {
	return types.GoType{Int8(a) / Int8(b)}
}

func divInt16(a, b types.Elem) types.Elem {
	return types.GoType{Int16(a) / Int16(b)}
}

func divInt32(a, b types.Elem) types.Elem {
	return types.GoType{Int32(a) / Int32(b)}
}

func divInt64(a, b types.Elem) types.Elem {
	return types.GoType{Int64(a) / Int64(b)}
}

func divUint(a, b types.Elem) types.Elem {
	return types.GoType{Uint(a) / Uint(b)}
}

func divUint8(a, b types.Elem) types.Elem {
	return types.GoType{Uint8(a) / Uint8(b)}
}

func divUint16(a, b types.Elem) types.Elem {
	return types.GoType{Uint16(a) / Uint16(b)}
}

func divUint32(a, b types.Elem) types.Elem {
	return types.GoType{Uint32(a) / Uint32(b)}
}

func divUint64(a, b types.Elem) types.Elem {
	return types.GoType{Uint64(a) / Uint64(b)}
}

func divFloat32(a, b types.Elem) types.Elem {
	return types.GoType{Float32(a) / Float32(b)}
}

func divFloat64(a, b types.Elem) types.Elem {
	return types.GoType{Float64(a) / Float64(b)}
}

func divComplex64(a, b types.Elem) types.Elem {
	return types.GoType{Complex64(a) / Complex64(b)}
}

func divComplex128(a, b types.Elem) types.Elem {
	return types.GoType{Complex128(a) / Complex128(b)}
}

func Div(s seq.Seq) types.Elem {
	var first, zero types.Elem

	if seq.Empty(s) {
		panic("Div cannot be called with no arguments")
	}
	zero, s, _ = s.FirstRest()
	first = zero

	var fn mathReduceFn
	switch first.(types.GoType).V.(type) {

	case int:
		fn = divInt

	case int8:
		fn = divInt8

	case int16:
		fn = divInt16

	case int32:
		fn = divInt32

	case int64:
		fn = divInt64

	case uint:
		fn = divUint

	case uint8:
		fn = divUint8

	case uint16:
		fn = divUint16

	case uint32:
		fn = divUint32

	case uint64:
		fn = divUint64

	case float32:
		fn = divFloat32

	case float64:
		fn = divFloat64

	case complex64:
		fn = divComplex64

	case complex128:
		fn = divComplex128

	default:
		panic(fmt.Sprintf("$#v cannot have Div called on it", first))
	}

	return mathReduce(fn, zero, s)
}
func modInt(a, b types.Elem) types.Elem {
	return types.GoType{Int(a) % Int(b)}
}

func modInt8(a, b types.Elem) types.Elem {
	return types.GoType{Int8(a) % Int8(b)}
}

func modInt16(a, b types.Elem) types.Elem {
	return types.GoType{Int16(a) % Int16(b)}
}

func modInt32(a, b types.Elem) types.Elem {
	return types.GoType{Int32(a) % Int32(b)}
}

func modInt64(a, b types.Elem) types.Elem {
	return types.GoType{Int64(a) % Int64(b)}
}

func modUint(a, b types.Elem) types.Elem {
	return types.GoType{Uint(a) % Uint(b)}
}

func modUint8(a, b types.Elem) types.Elem {
	return types.GoType{Uint8(a) % Uint8(b)}
}

func modUint16(a, b types.Elem) types.Elem {
	return types.GoType{Uint16(a) % Uint16(b)}
}

func modUint32(a, b types.Elem) types.Elem {
	return types.GoType{Uint32(a) % Uint32(b)}
}

func modUint64(a, b types.Elem) types.Elem {
	return types.GoType{Uint64(a) % Uint64(b)}
}

func Mod(s seq.Seq) types.Elem {
	var first, zero types.Elem

	if seq.Empty(s) {
		panic("Mod cannot be called with no arguments")
	}
	zero, s, _ = s.FirstRest()
	first = zero

	var fn mathReduceFn
	switch first.(types.GoType).V.(type) {

	case int:
		fn = modInt

	case int8:
		fn = modInt8

	case int16:
		fn = modInt16

	case int32:
		fn = modInt32

	case int64:
		fn = modInt64

	case uint:
		fn = modUint

	case uint8:
		fn = modUint8

	case uint16:
		fn = modUint16

	case uint32:
		fn = modUint32

	case uint64:
		fn = modUint64

	default:
		panic(fmt.Sprintf("$#v cannot have Mod called on it", first))
	}

	return mathReduce(fn, zero, s)
}
