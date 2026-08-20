// Package eventstudy measures what happened after each entry signal: how long
// the price took to reach a set of profit targets, and how far it fell first.
// It is the only place that knows the order of the whole analysis.
package eventstudy

import (
	"fmt"
	"time"
)

// Window is the slice of history the study runs on.
type Window struct {
	From, To int // inclusive indices into the full price history
	Years    int // years actually covered, for the report title
}

// ResolveWindow prefers a full `prefer`-year window and falls back to
// `fallback` years when the history is shorter.
//
// The fallback is not cosmetic: a ticker listed in 2010 has no 20-year window,
// and silently analysing 16 years while labelling it "20년" is the kind of
// mislabelling that makes a table wrong rather than merely incomplete. The
// caller gets the real number back and is expected to print it.
func ResolveWindow(dates []time.Time, prefer, fallback int) (Window, error) {
	if len(dates) == 0 {
		return Window{}, fmt.Errorf("no price data")
	}
	last := dates[len(dates)-1]

	for _, years := range []int{prefer, fallback} {
		if years <= 0 {
			continue
		}
		start := last.AddDate(-years, 0, 0)
		if !dates[0].After(start) {
			return Window{From: indexOnOrAfter(dates, start), To: len(dates) - 1, Years: years}, nil
		}
	}
	return Window{}, fmt.Errorf("history starts %s: not enough data for a %d-year window (fallback %d years)",
		dates[0].Format("2006-01-02"), prefer, fallback)
}

func indexOnOrAfter(dates []time.Time, t time.Time) int {
	for i, d := range dates {
		if !d.Before(t) {
			return i
		}
	}
	return len(dates) - 1
}
