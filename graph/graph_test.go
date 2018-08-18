package graph

import (
	"fmt"
	. "testing"
	"time"

	"github.com/mediocregopher/mediocre-go-lib/mrand"
	"github.com/mediocregopher/mediocre-go-lib/mtest/massert"
	"github.com/mediocregopher/mediocre-go-lib/mtest/mchk"
)

func strV(s string) Value {
	return Value{ID: s, V: s}
}

func TestGraph(t *T) {
	t.Parallel()
	type state struct {
		Graph

		m map[string]Edge
	}

	type params struct {
		add Edge
		del Edge
	}

	chk := mchk.Checker{
		Init: func() mchk.State {
			return state{
				m: map[string]Edge{},
			}
		},
		Next: func(ss mchk.State) mchk.Action {
			s := ss.(state)
			var p params
			if i := mrand.Intn(10); i == 0 {
				// add edge which is already there
				for _, e := range s.m {
					p.add = e
					break
				}
			} else if i == 1 {
				// delete edge which isn't there
				p.del = Edge{Tail: strV("z"), Head: strV("z")}
			} else if i <= 5 {
				// add probably new edge
				p.add = Edge{Tail: strV(mrand.Hex(1)), Head: strV(mrand.Hex(1))}
			} else {
				// probably del edge
				p.del = Edge{Tail: strV(mrand.Hex(1)), Head: strV(mrand.Hex(1))}
			}
			return mchk.Action{Params: p}
		},
		Apply: func(ss mchk.State, a mchk.Action) (mchk.State, error) {
			s, p := ss.(state), a.Params.(params)
			if p.add != (Edge{}) {
				s.Graph = s.Graph.AddEdge(p.add)
				s.m[p.add.id()] = p.add
			} else {
				s.Graph = s.Graph.DelEdge(p.del)
				delete(s.m, p.del.id())
			}

			{ // test Values and Edges methods
				vals := s.Graph.Values()
				edges := s.Graph.Edges()
				var aa []massert.Assertion
				found := map[string]bool{}
				tryAssert := func(v Value) {
					if ok := found[v.ID]; !ok {
						found[v.ID] = true
						aa = append(aa, massert.Has(vals, v))
					}
				}
				for _, e := range s.m {
					aa = append(aa, massert.Has(edges, e))
					tryAssert(e.Head)
					tryAssert(e.Tail)
				}
				aa = append(aa, massert.Len(vals, len(found)))
				aa = append(aa, massert.Len(edges, len(s.m)))
				if err := massert.All(aa...).Assert(); err != nil {
					return nil, err
				}
			}

			{ // test ValueEdges
				for _, val := range s.Graph.Values() {
					in, out := s.Graph.ValueEdges(val)
					var expIn, expOut []Edge
					for _, e := range s.m {
						if e.Tail.ID == val.ID {
							expOut = append(expOut, e)
						}
						if e.Head.ID == val.ID {
							expIn = append(expIn, e)
						}
					}
					if err := massert.Comment(massert.All(
						massert.Subset(expIn, in),
						massert.Len(in, len(expIn)),
						massert.Subset(expOut, out),
						massert.Len(out, len(expOut)),
					), "val:%q", val.V).Assert(); err != nil {
						return nil, err
					}
				}
			}

			return s, nil
		},
	}

	if err := chk.RunFor(5 * time.Second); err != nil {
		t.Fatal(err)
	}
}

func TestSubGraphAndEqual(t *T) {
	t.Parallel()
	type state struct {
		g1, g2                Graph
		expEqual, expSubGraph bool
	}

	type params struct {
		e          Edge
		add1, add2 bool
	}

	chk := mchk.Checker{
		Init: func() mchk.State {
			return state{expEqual: true, expSubGraph: true}
		},
		Next: func(ss mchk.State) mchk.Action {
			i := mrand.Intn(10)
			p := params{
				e:    Edge{Tail: strV(mrand.Hex(4)), Head: strV(mrand.Hex(4))},
				add1: i != 0,
				add2: i != 1,
			}
			return mchk.Action{Params: p}
		},
		Apply: func(ss mchk.State, a mchk.Action) (mchk.State, error) {
			s, p := ss.(state), a.Params.(params)
			if p.add1 {
				s.g1 = s.g1.AddEdge(p.e)
			}
			if p.add2 {
				s.g2 = s.g2.AddEdge(p.e)
			}
			s.expSubGraph = s.expSubGraph && p.add1
			s.expEqual = s.expEqual && p.add1 && p.add2

			if s.g1.SubGraph(s.g2) != s.expSubGraph {
				return nil, fmt.Errorf("SubGraph expected to return %v", s.expSubGraph)
			}

			if s.g1.Equal(s.g2) != s.expEqual {
				return nil, fmt.Errorf("Equal expected to return %v", s.expEqual)
			}

			return s, nil
		},
		MaxLength: 100,
	}

	if err := chk.RunFor(5 * time.Second); err != nil {
		t.Fatal(err)
	}
}
