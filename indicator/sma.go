// Package indicator computes technical indicators from a price series. Every
// function here is pure: no I/O, no state, no knowledge of where the prices
// came from. Warm-up positions are filled with NaN so callers can tell "not
// enough history yet" from a real value.
package indicator

import "math"

// SMA returns the simple moving average over window periods. The first
// window-1 positions are NaN.
//
// Feed this the full price history, not just the analysis window: slicing
// first would leave the first window-1 days of the window itself unusable.
func SMA(values []float64, window int) []float64 {
	out := make([]float64, len(values))
	if window <= 0 {
		for i := range out {
			out[i] = math.NaN()
		}
		return out
	}

	var sum float64
	for i, v := range values {
		sum += v
		if i >= window {
			sum -= values[i-window]
		}
		if i >= window-1 {
			out[i] = sum / float64(window)
		} else {
			out[i] = math.NaN()
		}
	}
	return out
}
