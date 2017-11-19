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

	assert.Equal(t, 2, r.Edge(Up))
	assert.Equal(t, 3, r.Edge(Down))
	assert.Equal(t, 1, r.Edge(Left))
	assert.Equal(t, 2, r.Edge(Right))

	assert.Equal(t, XY{1, 2}, r.Corner(Left, Up))
	assert.Equal(t, XY{1, 3}, r.Corner(Left, Down))
	assert.Equal(t, XY{2, 2}, r.Corner(Right, Up))
	assert.Equal(t, XY{2, 3}, r.Corner(Right, Down))
}

func TestRectCenter(t *T) {
	assertCentered := func(exp, given Rect, center XY, rounder Rounder) {
		got := given.Centered(center, rounder)
		assert.Equal(t, exp, got)
		assert.Equal(t, center, got.Center(rounder))
	}

	{
		r := Rect{
			Size: XY{4, 4},
		}
		assert.Equal(t, XY{2, 2}, r.Center(Round))
		assert.Equal(t, XY{2, 2}, r.Center(Floor))
		assert.Equal(t, XY{2, 2}, r.Center(Ceil))
		assertCentered(
			Rect{TopLeft: XY{1, 1}, Size: XY{4, 4}},
			r, XY{3, 3}, Round,
		)
		assertCentered(
			Rect{TopLeft: XY{1, 1}, Size: XY{4, 4}},
			r, XY{3, 3}, Floor,
		)
		assertCentered(
			Rect{TopLeft: XY{1, 1}, Size: XY{4, 4}},
			r, XY{3, 3}, Ceil,
		)
	}

	{
		r := Rect{
			Size: XY{5, 5},
		}
		assert.Equal(t, XY{3, 3}, r.Center(Round))
		assert.Equal(t, XY{2, 2}, r.Center(Floor))
		assert.Equal(t, XY{3, 3}, r.Center(Ceil))
		assertCentered(
			Rect{TopLeft: XY{0, 0}, Size: XY{5, 5}},
			r, XY{3, 3}, Round,
		)
		assertCentered(
			Rect{TopLeft: XY{1, 1}, Size: XY{5, 5}},
			r, XY{3, 3}, Floor,
		)
		assertCentered(
			Rect{TopLeft: XY{0, 0}, Size: XY{5, 5}},
			r, XY{3, 3}, Ceil,
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
