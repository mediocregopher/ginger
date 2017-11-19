package geo

import (
	"fmt"
	"math"
)

// Rounder describes how a floating point number should be converted to an int
type Rounder int

const (
	// Round will round up or down depending on the number itself
	Round Rounder = iota

	// Floor will use the math.Floor function
	Floor

	// Ceil will use the math.Ceil function
	Ceil
)

// Round64 converts a float to an in64 based on the rounding function indicated
// by the Rounder's value
func (r Rounder) Round64(f float64) int64 {
	switch r {
	case Round:
		if f < 0 {
			f = math.Ceil(f - 0.5)
		}
		f = math.Floor(f + 0.5)
	case Floor:
		f = math.Floor(f)
	case Ceil:
		f = math.Ceil(f)
	default:
		panic(fmt.Sprintf("invalid Rounder: %#v", r))
	}
	return int64(f)
}

// Round is like Round64 but convers the int64 to an int
func (r Rounder) Round(f float64) int {
	return int(r.Round64(f))
}
