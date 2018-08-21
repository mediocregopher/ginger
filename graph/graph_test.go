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
				s.Graph = s.Graph.Add(p.add)
				s.m[p.add.id()] = p.add
			} else {
				s.Graph = s.Graph.Del(p.del)
				delete(s.m, p.del.id())
			}

			{ // test Nodes and Edges methods
				nodes := s.Graph.Nodes()
				edges := s.Graph.Edges()
				var aa []massert.Assertion
				vals := map[string]bool{}
				ins, outs := map[string]int{}, map[string]int{}
				for _, e := range s.m {
					aa = append(aa, massert.Has(edges, e))
					aa = append(aa, massert.HasKey(nodes, e.Head.ID))
					aa = append(aa, massert.Has(nodes[e.Head.ID].Ins, e))
					aa = append(aa, massert.HasKey(nodes, e.Tail.ID))
					aa = append(aa, massert.Has(nodes[e.Tail.ID].Outs, e))
					vals[e.Head.ID] = true
					vals[e.Tail.ID] = true
					ins[e.Head.ID]++
					outs[e.Tail.ID]++
				}
				aa = append(aa, massert.Len(edges, len(s.m)))
				aa = append(aa, massert.Len(nodes, len(vals)))
				for id, node := range nodes {
					aa = append(aa, massert.Len(node.Ins, ins[id]))
					aa = append(aa, massert.Len(node.Outs, outs[id]))
				}

				if err := massert.All(aa...).Assert(); err != nil {
					return nil, err
				}
			}

			{ // test Node and Has. Nodes has already been tested so we can use
				// its returned Nodes as the expected ones
				var aa []massert.Assertion
				for _, expNode := range s.Graph.Nodes() {
					var naa []massert.Assertion
					node, ok := s.Graph.Node(expNode.Value)
					naa = append(naa, massert.Equal(true, ok))
					naa = append(naa, massert.Equal(true, s.Graph.Has(expNode.Value)))
					naa = append(naa, massert.Subset(expNode.Ins, node.Ins))
					naa = append(naa, massert.Len(node.Ins, len(expNode.Ins)))
					naa = append(naa, massert.Subset(expNode.Outs, node.Outs))
					naa = append(naa, massert.Len(node.Outs, len(expNode.Outs)))

					aa = append(aa, massert.Comment(massert.All(naa...), "v:%q", expNode.ID))
				}
				_, ok := s.Graph.Node(strV("zz"))
				aa = append(aa, massert.Equal(false, ok))
				aa = append(aa, massert.Equal(false, s.Graph.Has(strV("zz"))))

				if err := massert.All(aa...).Assert(); err != nil {
					return nil, err
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
				s.g1 = s.g1.Add(p.e)
			}
			if p.add2 {
				s.g2 = s.g2.Add(p.e)
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
