// Package stats provides the summary statistics shared by the event study and
// the valuation report. It knows nothing about prices, entries or PER/PBR — it
// only summarises a slice of numbers.
package stats

import (
	"math"
	"sort"
)

// Summary describes a sample. N is the number of values actually summarised;
// callers that drop censored observations must report that separately, because
// a smaller N with no explanation is how a right-censored sample gets misread.
type Summary struct {
	N      int
	Mean   float64
	Median float64
	Min    float64
	Max    float64
}

// Summarize returns the summary of values. NaN entries are ignored. An empty
// sample yields the zero Summary, so check N before formatting.
func Summarize(values []float64) Summary {
	clean := make([]float64, 0, len(values))
	for _, v := range values {
		if !math.IsNaN(v) {
			clean = append(clean, v)
		}
	}
	if len(clean) == 0 {
		return Summary{}
	}
	sort.Float64s(clean)

	var sum float64
	for _, v := range clean {
		sum += v
	}

	return Summary{
		N:      len(clean),
		Mean:   sum / float64(len(clean)),
		Median: median(clean),
		Min:    clean[0],
		Max:    clean[len(clean)-1],
	}
}

// median expects a sorted, non-empty slice.
func median(sorted []float64) float64 {
	mid := len(sorted) / 2
	if len(sorted)%2 == 1 {
		return sorted[mid]
	}
	return (sorted[mid-1] + sorted[mid]) / 2
}
