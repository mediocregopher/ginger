package main

import (
	. "testing"

	"github.com/mediocregopher/ginger/gim/geo"
	"github.com/stretchr/testify/assert"
)

func TestBox(t *T) {
	b := box{
		pos:  geo.XY{1, 2},
		size: geo.XY{10, 11},
	}

	assert.Equal(t, geo.XY{10, 11}, b.innerSize())
	assert.Equal(t, geo.XY{12, 13}, b.rectSize())

	assert.Equal(t, 2, b.rectEdge(geo.Up))
	assert.Equal(t, 15, b.rectEdge(geo.Down))
	assert.Equal(t, 1, b.rectEdge(geo.Left))
	assert.Equal(t, 13, b.rectEdge(geo.Right))

	assert.Equal(t, geo.XY{1, 2}, b.rectCorner(geo.Left, geo.Up))
	assert.Equal(t, geo.XY{1, 14}, b.rectCorner(geo.Left, geo.Down))
	assert.Equal(t, geo.XY{12, 2}, b.rectCorner(geo.Right, geo.Up))
	assert.Equal(t, geo.XY{12, 14}, b.rectCorner(geo.Right, geo.Down))
}
