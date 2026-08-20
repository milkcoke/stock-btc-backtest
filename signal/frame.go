// Package signal decides when a price series meets an entry rule and turns
// those days into discrete buy events. It computes no indicators of its own —
// it reads a Frame that has already been prepared.
package signal

import (
	"math"
	"time"
)

// Frame is a read-only, index-aligned view of everything a condition may need.
//
// Passing the whole frame rather than individual numbers is what keeps
// Condition open for extension: a new condition can reach for a series the
// existing ones ignore without changing the interface or the detector.
type Frame struct {
	Dates  []time.Time
	Prices []float64
	SMA    []float64
	RSI    []float64
	Peak   []float64 // running high, reset each calendar year
}

func (f Frame) Len() int { return len(f.Prices) }

// DrawdownPct is the percentage below the running peak, negative when under it.
func (f Frame) DrawdownPct(i int) float64 {
	if f.Peak[i] <= 0 {
		return math.NaN()
	}
	return (f.Prices[i]/f.Peak[i] - 1) * 100
}

// MADiscountPct is the percentage below the moving average, negative when under it.
func (f Frame) MADiscountPct(i int) float64 {
	if math.IsNaN(f.SMA[i]) || f.SMA[i] <= 0 {
		return math.NaN()
	}
	return (f.Prices[i]/f.SMA[i] - 1) * 100
}

// Ready reports whether every indicator has finished warming up at i. Conditions
// need not check this themselves; the detector skips days that are not ready.
func (f Frame) Ready(i int) bool {
	return !math.IsNaN(f.SMA[i]) && !math.IsNaN(f.RSI[i]) && f.Peak[i] > 0
}
