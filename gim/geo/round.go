package geo

import (
	"math"
)

// RounderFunc is a function which converts a floating point number into an
// integer.
type RounderFunc func(float64) int64

// Round is helper for calling the RounderFunc and converting the result to an
// int.
func (rf RounderFunc) Round(f float64) int {
	return int(rf(f))
}

// A few RounderFuncs which can be used. Set the Rounder global variable to pick
// one.
var (
	Floor RounderFunc = func(f float64) int64 { return int64(math.Floor(f)) }
	Ceil  RounderFunc = func(f float64) int64 { return int64(math.Ceil(f)) }
	Round RounderFunc = func(f float64) int64 {
		if f < 0 {
			f = math.Ceil(f - 0.5)
		}
		f = math.Floor(f + 0.5)
		return int64(f)
	}
)

// Rounder is the RounderFunc which will be used by all functions and methods in
// this package when needed.
var Rounder = Ceil
