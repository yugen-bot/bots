package utils

import "math"

// RoundTwo rounds f to two decimal places to avoid float64 drift when
// accumulating fractional vote rewards (e.g. 0.25 + 0.2).
func RoundTwo(f float64) float64 {
	return math.Round(f*100) / 100
}
