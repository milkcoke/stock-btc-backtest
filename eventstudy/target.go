package eventstudy

import "time"

// Outcome records whether one entry reached one profit target.
type Outcome struct {
	Target   float64 // percent, e.g. 20
	Achieved bool
	Days     int // calendar days to the hit, or days elapsed so far when not achieved
	HitIndex int // index of the hit, or the last observed day when not achieved
	HitDate  time.Time
	HitPrice float64
}

// TargetTracker measures the time from a buy to the first close at or above
// each target.
type TargetTracker struct{ Targets []float64 }

// Track returns one Outcome per target, in the order given.
//
// An entry that has not reached its target yet is returned with Achieved=false
// and Days counted up to the end of the data. That distinction has to survive
// all the way to the report: a position opened three weeks ago is not a slow
// case, it is an unfinished one, and dropping it silently pulls the fastest
// observation out of the sample and inflates every average above it.
func (t TargetTracker) Track(prices []float64, dates []time.Time, buyIndex, to int) []Outcome {
	buy, buyDate := prices[buyIndex], dates[buyIndex]
	out := make([]Outcome, 0, len(t.Targets))

	for _, target := range t.Targets {
		goal := buy * (1 + target/100)
		o := Outcome{Target: target, Days: daysBetween(buyDate, dates[to]), HitIndex: to}
		for i := buyIndex + 1; i <= to; i++ {
			if prices[i] >= goal {
				o = Outcome{Target: target, Achieved: true, Days: daysBetween(buyDate, dates[i]),
					HitIndex: i, HitDate: dates[i], HitPrice: prices[i]}
				break
			}
		}
		out = append(out, o)
	}
	return out
}

// LowestUntil returns the deepest close between the buy and `until`, which is
// what the position actually had to sit through.
func LowestUntil(prices []float64, dates []time.Time, buyIndex, until int) (time.Time, float64) {
	low, lowAt := prices[buyIndex], buyIndex
	for i := buyIndex; i <= until && i < len(prices); i++ {
		if prices[i] < low {
			low, lowAt = prices[i], i
		}
	}
	return dates[lowAt], low
}

func daysBetween(a, b time.Time) int {
	return int(b.Sub(a).Hours() / 24)
}
