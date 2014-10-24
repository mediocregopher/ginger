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
{{range .}}{{$mathOp := .}}{{range .OpTypes}}
func {{$mathOp.Private}}{{.CastFn}}(a, b types.Elem) types.Elem {
	return types.GoType{{`{`}}{{.CastFn}}(a) {{$mathOp.Op}} {{.CastFn}}(b)}
}
{{end}}{{if $mathOp.IncludeString}}
func {{$mathOp.Private}}String(a, b types.Elem) types.Elem {
	return types.GoType{String(a) {{$mathOp.Op}} String(b)}
}
{{end}}
func {{$mathOp.Public}}(s seq.Seq) types.Elem {
	var first, zero types.Elem
{{if (eq $mathOp.Unit "")}}
	if seq.Empty(s) {
		panic("{{$mathOp.Public}} cannot be called with no arguments")
	}
	zero, s, _ = s.FirstRest()
	first = zero
{{else}}
	if seq.Empty(s) {
		return types.GoType{{`{`}}{{$mathOp.Unit}}}
	}
	first, _, _ = s.FirstRest()
{{end}}
	var fn mathReduceFn
	switch first.(types.GoType).V.(type) {
{{range .OpTypes}}
	case {{.Type}}:
		fn = {{$mathOp.Private}}{{.CastFn}}{{if (ne $mathOp.Unit "")}}
		zero = types.GoType{{`{`}}{{.Type}}({{$mathOp.Unit}})}{{end}}
{{end}}{{if $mathOp.IncludeString}}
	case string:
		fn = {{$mathOp.Private}}String
		zero = types.GoType{string("")}
{{end}}
	default:
		panic(fmt.Sprintf("$#v cannot have {{$mathOp.Public}} called on it", first))
	}

	return mathReduce(fn, zero, s)
}{{end}}
