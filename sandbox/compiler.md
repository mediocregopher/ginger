I need to figure out how the compiler-time vs run-time execution is going to
work, and how I'm going to differentiate between the two in the language.

main := MainFunc()
foo := main.Int(1)

incrFunc := main.NewFunction(inType, outType)
in := incrFunc.In()
add := incrFunc.Var("add") // should be macro?
out := incrFunc.Call(add, incrFunc.Int(1), in)
incrFunc.Return(out) // ugly

main.Return(main.Call(incrFunc, foo))

compiler := NewCompiler()
compiler.Enter(main)

////////////////////////////////////////////////////////////////////////////////

type val { type, llvmVal }

type func { type, llvmVal }

////////////////////////////////////////////////////////////////////////////////

MACRO DISPATCHER as the thing which has a set of exposed methods. defmacro like
thing can be built on top of it.

TYPED HEAP. Kind of like a typed map mixed with a set. Maybe looks like

```
h := make(heap[float64], 10)
id := h.add(8.5)
eightPointFive := h.get(id)
h.del(id)
```

Since the heap is a known size and each element in it is as well it can be
statically allocated at one spot in the stack and the pointer to it passed
farther into the stack as needed.
