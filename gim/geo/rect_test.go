package geo

import (
	. "testing"

	"github.com/stretchr/testify/assert"
)

func TestRect(t *T) {
	r := Rect{
		TopLeft: XY{1, 2},
		Size:    XY{2, 2},
	}

	assert.Equal(t, 2, r.EdgeCoord(Up))
	assert.Equal(t, 3, r.EdgeCoord(Down))
	assert.Equal(t, 1, r.EdgeCoord(Left))
	assert.Equal(t, 2, r.EdgeCoord(Right))

	lu := XY{1, 2}
	ld := XY{1, 3}
	ru := XY{2, 2}
	rd := XY{2, 3}

	assert.Equal(t, lu, r.Corner(Left, Up))
	assert.Equal(t, ld, r.Corner(Left, Down))
	assert.Equal(t, ru, r.Corner(Right, Up))
	assert.Equal(t, rd, r.Corner(Right, Down))

	assert.Equal(t, Edge{lu, ld}, r.Edge(Left, Down))
	assert.Equal(t, Edge{ru, rd}, r.Edge(Right, Down))
	assert.Equal(t, Edge{lu, ru}, r.Edge(Up, Right))
	assert.Equal(t, Edge{ld, rd}, r.Edge(Down, Right))
	assert.Equal(t, Edge{ld, lu}, r.Edge(Left, Up))
	assert.Equal(t, Edge{rd, ru}, r.Edge(Right, Up))
	assert.Equal(t, Edge{ru, lu}, r.Edge(Up, Left))
	assert.Equal(t, Edge{rd, ld}, r.Edge(Down, Left))
}

func TestRectCenter(t *T) {
	assertCentered := func(exp, given Rect, center XY) {
		got := given.Centered(center)
		assert.Equal(t, exp, got)
		assert.Equal(t, center, got.Center())
	}

	{
		r := Rect{
			Size: XY{4, 4},
		}
		assert.Equal(t, XY{2, 2}, r.Center())
		assertCentered(
			Rect{TopLeft: XY{1, 1}, Size: XY{4, 4}},
			r, XY{3, 3},
		)
	}

	{
		r := Rect{
			Size: XY{5, 5},
		}
		assert.Equal(t, XY{3, 3}, r.Center())
		assertCentered(
			Rect{TopLeft: XY{0, 0}, Size: XY{5, 5}},
			r, XY{3, 3},
		)
	}
}

func TestRectUnion(t *T) {
	assertUnion := func(exp, r1, r2 Rect) {
		assert.Equal(t, exp, r1.Union(r2))
		assert.Equal(t, exp, r2.Union(r1))
	}

	{ // Zero
		r := Rect{TopLeft: XY{1, 1}, Size: XY{2, 2}}
		assertUnion(r, r, Rect{})
	}

	{ // Equal
		r := Rect{Size: XY{2, 2}}
		assertUnion(r, r, r)
	}

	{ // Overlapping corner
		r1 := Rect{TopLeft: XY{0, 0}, Size: XY{2, 2}}
		r2 := Rect{TopLeft: XY{1, 1}, Size: XY{2, 2}}
		ex := Rect{TopLeft: XY{0, 0}, Size: XY{3, 3}}
		assertUnion(ex, r1, r2)
	}

	{ // 2 overlapping corners
		r1 := Rect{TopLeft: XY{0, 0}, Size: XY{4, 4}}
		r2 := Rect{TopLeft: XY{1, 1}, Size: XY{4, 2}}
		ex := Rect{TopLeft: XY{0, 0}, Size: XY{5, 4}}
		assertUnion(ex, r1, r2)
	}

	{ // Shared edge
		r1 := Rect{TopLeft: XY{0, 0}, Size: XY{2, 1}}
		r2 := Rect{TopLeft: XY{1, 0}, Size: XY{1, 2}}
		ex := Rect{TopLeft: XY{0, 0}, Size: XY{2, 2}}
		assertUnion(ex, r1, r2)
	}

	{ // Adjacent edge
		r1 := Rect{TopLeft: XY{0, 0}, Size: XY{2, 2}}
		r2 := Rect{TopLeft: XY{2, 0}, Size: XY{2, 2}}
		ex := Rect{TopLeft: XY{0, 0}, Size: XY{4, 2}}
		assertUnion(ex, r1, r2)
	}
}
