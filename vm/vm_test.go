package vm

import (
	. "testing"

	"github.com/mediocregopher/ginger/lang"
)

func TestCompiler(t *T) {
	mkcmd := func(a lang.Atom, args ...lang.Term) lang.Tuple {
		// TODO a 1-tuple should be the same as its element?
		if len(args) == 1 {
			return lang.Tuple{a, args[0]}
		}
		return lang.Tuple{a, lang.Tuple(args)}
	}
	mkint := func(i string) lang.Tuple {
		return lang.Tuple{Int, lang.Const(i)}
	}

	type test struct {
		in  []lang.Term
		exp uint64
	}

	one := mkint("1")
	two := mkint("2")
	foo := mkcmd(Var, lang.Atom("foo"))
	bar := mkcmd(Var, lang.Atom("bar"))
	baz := mkcmd(Var, lang.Atom("baz"))

	tests := []test{
		{
			in:  []lang.Term{one},
			exp: 1,
		},
		{
			in: []lang.Term{
				mkcmd(Add, mkcmd(Tuple, one, two)),
			},
			exp: 3,
		},
		{
			in: []lang.Term{
				mkcmd(Assign, foo, one),
				mkcmd(Add, mkcmd(Tuple, foo, two)),
			},
			exp: 3,
		},
		{
			in: []lang.Term{
				mkcmd(Assign, foo, mkcmd(Tuple, one, two)),
				mkcmd(Add, foo),
			},
			exp: 3,
		},
		{
			in: []lang.Term{
				mkcmd(Assign, foo, mkcmd(Tuple, one, two)),
				mkcmd(Assign, bar, mkcmd(Add, foo)),
				mkcmd(Assign, baz, mkcmd(Add, foo)),
				mkcmd(Add, mkcmd(Tuple, bar, baz)),
			},
			exp: 6,
		},
	}

	for _, test := range tests {
		t.Logf("testing program: %v", test.in)
		mod, err := Build(test.in...)
		if err != nil {
			t.Fatalf("building failed: %s", err)
		}

		out, err := mod.Run()
		if err != nil {
			mod.Dump()
			t.Fatalf("running failed: %s", err)
		} else if out != test.exp {
			mod.Dump()
			t.Fatalf("expected result %T:%v, got %T:%v", test.exp, test.exp, out, out)
		}
	}
}
