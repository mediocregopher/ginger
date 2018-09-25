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

func TestDisjoinUnion(t *T) {
	t.Parallel()
	type state struct {
		g Graph
		// prefix -> Values with that prefix. contains dupes
		valM  map[string][]Value
		disjM map[string]Graph
	}

	type params struct {
		prefix string
		e      Edge
	}

	chk := mchk.Checker{
		Init: func() mchk.State {
			return state{
				valM:  map[string][]Value{},
				disjM: map[string]Graph{},
			}
		},
		Next: func(ss mchk.State) mchk.Action {
			s := ss.(state)
			prefix := mrand.Hex(1)
			var edge Edge
			if vals := s.valM[prefix]; len(vals) == 0 {
				edge = Edge{
					Tail: strV(prefix + mrand.Hex(1)),
					Head: strV(prefix + mrand.Hex(1)),
				}
			} else if mrand.Intn(2) == 0 {
				edge = Edge{
					Tail: mrand.Element(vals, nil).(Value),
					Head: strV(prefix + mrand.Hex(1)),
				}
			} else {
				edge = Edge{
					Tail: strV(prefix + mrand.Hex(1)),
					Head: mrand.Element(vals, nil).(Value),
				}
			}

			return mchk.Action{Params: params{prefix: prefix, e: edge}}
		},
		Apply: func(ss mchk.State, a mchk.Action) (mchk.State, error) {
			s, p := ss.(state), a.Params.(params)
			s.g = s.g.Add(p.e)
			s.valM[p.prefix] = append(s.valM[p.prefix], p.e.Head, p.e.Tail)
			s.disjM[p.prefix] = s.disjM[p.prefix].Add(p.e)

			var aa []massert.Assertion

			// test Disjoin
			disj := s.g.Disjoin()
			for prefix, graph := range s.disjM {
				aa = append(aa, massert.Comment(
					massert.Equal(true, graph.Equal(s.disjM[prefix])),
					"prefix:%q", prefix,
				))
			}
			aa = append(aa, massert.Len(disj, len(s.disjM)))

			// now test Join
			join := (Graph{}).Join(disj...)
			aa = append(aa, massert.Equal(true, s.g.Equal(join)))

			return s, massert.All(aa...).Assert()
		},
		MaxLength: 100,
		// Each action is required for subsequent ones to make sense, so
		// minimizing won't work
		DontMinimize: true,
	}

	if err := chk.RunFor(5 * time.Second); err != nil {
		t.Fatal(err)
	}
}

func TestVisitBreadth(t *T) {
	t.Parallel()
	type state struct {
		g Graph
		// each rank describes the set of values (by ID) which should be
		// visited in that rank. Within a rank the values will be visited in any
		// order
		ranks []map[string]bool
	}

	thisRank := func(s state) map[string]bool {
		return s.ranks[len(s.ranks)-1]
	}

	prevRank := func(s state) map[string]bool {
		return s.ranks[len(s.ranks)-2]
	}

	randFromRank := func(s state, rankPickFn func(state) map[string]bool) Value {
		rank := rankPickFn(s)
		rankL := make([]string, 0, len(rank))
		for id := range rank {
			rankL = append(rankL, id)
		}
		return strV(mrand.Element(rankL, nil).(string))
	}

	type params struct {
		newRank bool
		e       Edge
	}

	chk := mchk.Checker{
		Init: func() mchk.State {
			return state{
				ranks: []map[string]bool{
					map[string]bool{"start": true},
					map[string]bool{},
				},
			}
		},
		Next: func(ss mchk.State) mchk.Action {
			s := ss.(state)
			var p params
			p.newRank = len(thisRank(s)) > 0 && mrand.Intn(10) == 0
			if p.newRank {
				p.e.Head = strV(mrand.Hex(3))
				p.e.Tail = randFromRank(s, thisRank)
			} else {
				p.e.Head = strV(mrand.Hex(2))
				p.e.Tail = randFromRank(s, prevRank)
			}
			return mchk.Action{Params: p}
		},
		Apply: func(ss mchk.State, a mchk.Action) (mchk.State, error) {
			s, p := ss.(state), a.Params.(params)
			if p.newRank {
				s.ranks = append(s.ranks, map[string]bool{})
			}
			if !s.g.Has(p.e.Head) {
				thisRank(s)[p.e.Head.ID] = true
			}
			s.g = s.g.Add(p.e)

			// check the visit
			var err error
			expRanks := s.ranks
			currRank := map[string]bool{}
			s.g.VisitBreadth(strV("start"), func(n Node) bool {
				currRank[n.Value.ID] = true
				if len(currRank) != len(expRanks[0]) {
					return true
				}
				if err = massert.Equal(expRanks[0], currRank).Assert(); err != nil {
					return false
				}
				expRanks = expRanks[1:]
				currRank = map[string]bool{}
				return true
			})
			if err != nil {
				return nil, err
			}

			err = massert.All(
				massert.Len(expRanks, 0),
				massert.Len(currRank, 0),
			).Assert()
			return s, err
		},
		DontMinimize: true,
	}

	if err := chk.RunCase(); err != nil {
		t.Fatal(err)
	}

	if err := chk.RunFor(5 * time.Second); err != nil {
		t.Fatal(err)
	}
}
