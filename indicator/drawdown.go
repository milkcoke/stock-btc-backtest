package indicator

import "time"

// RunningPeak returns the highest value seen so far at each position, reset at
// the start of every calendar year.
//
// The yearly reset matters: a rule written against "the year's high" cannot be
// executed, because the year's high is only known in hindsight. The running
// peak is what an investor actually sees on the day.
func RunningPeak(dates []time.Time, values []float64) []float64 {
	out := make([]float64, len(values))
	peak, year := 0.0, 0
	for i, v := range values {
		if y := dates[i].Year(); y != year {
			year, peak = y, 0
		}
		if v > peak {
			peak = v
		}
		out[i] = peak
	}
	return out
}

// YearMDD is the deepest drawdown inside one calendar year, measured from the
// running peak at the time — not from the year's opening price.
type YearMDD struct {
	Year       int
	Pct        float64 // negative, e.g. -30.2
	Peak       float64
	PeakDate   time.Time
	Trough     float64
	TroughDate time.Time
	Complete   bool // false when the year is only partly covered by the data
	Days       int  // trading days observed in this year
}

// YearlyMDD returns one entry per calendar year present in the data.
//
// Complete is false for the first and last year whenever the data does not
// cover them from January to December. Partial years understate the drawdown,
// so averaging them in would bias the threshold towards zero; callers should
// filter on Complete before taking a mean.
func YearlyMDD(dates []time.Time, values []float64) []YearMDD {
	if len(values) == 0 {
		return nil
	}
	peaks := RunningPeak(dates, values)

	var out []YearMDD
	var cur *YearMDD
	for i, v := range values {
		y := dates[i].Year()
		if cur == nil || cur.Year != y {
			out = append(out, YearMDD{Year: y, Peak: peaks[i], PeakDate: dates[i],
				Trough: v, TroughDate: dates[i]})
			cur = &out[len(out)-1]
		}
		cur.Days++
		if dd := (v/peaks[i] - 1) * 100; dd < cur.Pct {
			cur.Pct = dd
			cur.Peak, cur.Trough, cur.TroughDate = peaks[i], v, dates[i]
			cur.PeakDate = peakDateFor(dates, values, i, peaks[i])
		}
	}

	first, last := dates[0], dates[len(dates)-1]
	for i := range out {
		out[i].Complete = true
		if out[i].Year == first.Year() && !(first.Month() == time.January && first.Day() <= 7) {
			out[i].Complete = false
		}
		if out[i].Year == last.Year() && !(last.Month() == time.December && last.Day() >= 24) {
			out[i].Complete = false
		}
	}
	return out
}

// peakDateFor walks back to the day the running peak was set, so the report can
// name both ends of the drawdown.
func peakDateFor(dates []time.Time, values []float64, from int, peak float64) time.Time {
	year := dates[from].Year()
	for i := from; i >= 0 && dates[i].Year() == year; i-- {
		if values[i] == peak {
			return dates[i]
		}
	}
	return dates[from]
}
