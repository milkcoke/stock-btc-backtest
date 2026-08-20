package signal

import "time"

// Entry is one buy event: the day the rule first fired in a given decline.
type Entry struct {
	Index     int
	BuyDate   time.Time
	Buy       float64
	PeakDate  time.Time
	Peak      float64
	VsPeakPct float64
	VsMAPct   float64
	RSI       float64
}

// Detector turns rule-satisfying days into discrete entries. Its only
// responsibility is the re-arm policy — it never inspects the rule itself.
type Detector struct{ Condition Condition }

// Detect scans [from, to] and returns at most one entry per decline.
//
// After firing, the detector disarms until the price sets a new running peak.
// Without that, a single bear market would produce dozens of "entries" on
// consecutive days and every statistic downstream would count the same decline
// many times over.
//
// Note what this assumes: the whole position is bought on the first qualifying
// day, which is the worst possible fill within that decline. Days that keep
// meeting the rule afterwards — often dozens, at prices far below the first —
// are ignored. Results are therefore conservative, and any report built on them
// should say so.
func (d Detector) Detect(f Frame, from, to int) []Entry {
	var entries []Entry
	armed := true
	peak, peakDate, year := 0.0, time.Time{}, 0

	for i := from; i <= to && i < f.Len(); i++ {
		if y := f.Dates[i].Year(); y != year {
			year, peak, peakDate, armed = y, 0, time.Time{}, true
		}

		if f.Prices[i] > peak {
			peak, peakDate, armed = f.Prices[i], f.Dates[i], true
			continue
		}
		if !armed || !f.Ready(i) || !d.Condition.Met(f, i) {
			continue
		}

		entries = append(entries, Entry{
			Index:     i,
			BuyDate:   f.Dates[i],
			Buy:       f.Prices[i],
			PeakDate:  peakDate,
			Peak:      peak,
			VsPeakPct: f.DrawdownPct(i),
			VsMAPct:   f.MADiscountPct(i),
			RSI:       f.RSI[i],
		})
		armed = false
	}
	return entries
}
