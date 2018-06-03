package terminal

import (
	"fmt"
	"math/rand"
	"strings"
	. "testing"
	"time"
)

func TestMat(t *T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	type xy struct {
		x, y int
	}

	type action struct {
		xy
		set int
	}

	run := func(aa []action) {
		aaStr := func(i int) string {
			s := fmt.Sprintf("%#v", aa[:i+1])
			return strings.Replace(s, "terminal.", "", -1)
		}

		m := newMat()
		mm := map[xy]int{}
		for i, a := range aa {
			if a.set > 0 {
				mm[a.xy] = a.set
				m.set(a.xy.x, a.xy.y, a.set)
				continue
			}

			expI, expOk := mm[a.xy]
			gotI, gotOk := m.get(a.xy.x, a.xy.y).(int)
			if expOk != gotOk {
				t.Fatalf("get failed: expOk:%v gotOk:%v actions:%#v", expOk, gotOk, aaStr(i))
			} else if expI != gotI {
				t.Fatalf("get failed: expI:%v gotI:%v actions:%#v", expI, gotI, aaStr(i))
			}
		}
	}

	for i := 0; i < 10000; i++ {
		var actions []action
		for j := r.Intn(1000); j > 0; j-- {
			a := action{xy: xy{x: r.Intn(5), y: r.Intn(5)}}
			if r.Intn(3) == 0 {
				a.set = r.Intn(10000) + 1
			}
			actions = append(actions, a)
		}
		run(actions)
	}
}
