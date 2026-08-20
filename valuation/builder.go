package valuation

import (
	"fmt"
	"sort"
	"time"

	"stock-btc-backtest/loader"
)

// Builder derives a daily PER/PBR series from reported fundamentals and a price
// series.
//
//	PER = close ÷ trailing-twelve-month EPS
//	PBR = close ÷ (stockholders' equity ÷ shares outstanding)
//
// Both sources are injected so the derivation can be tested, and so a different
// filing system (DART for Korean issuers, say) can be plugged in without
// touching the arithmetic.
type Builder struct {
	Fundamentals FundamentalsSource
	Splits       SplitSource
}

// Build returns one record per trading day that has both a published earnings
// figure and a published book value behind it.
//
// Days where trailing earnings are zero or negative are dropped rather than
// reported as a huge or negative PER, which is the same convention exchanges
// use — a P/E on a loss is not a number anyone can act on.
func (b Builder) Build(ticker string, prices []loader.PriceRecord) ([]Record, error) {
	fundamentals, err := b.Fundamentals.Fetch(ticker)
	if err != nil {
		return nil, err
	}
	splits, err := b.Splits.Splits(ticker)
	if err != nil {
		return nil, fmt.Errorf("%s: split history: %w", ticker, err)
	}

	eps := trailingEPS(fundamentals, splits)
	bps := bookValuePerShare(fundamentals, splits)
	if len(eps) == 0 || len(bps) == 0 {
		return nil, fmt.Errorf("%s: not enough reported history to derive PER/PBR", ticker)
	}

	var out []Record
	epsAt, bpsAt := newCursor(eps), newCursor(bps)
	for _, p := range prices {
		e, okE := epsAt.at(p.Date)
		v, okV := bpsAt.at(p.Date)
		if !okE || !okV || e <= 0 || v <= 0 {
			continue
		}
		out = append(out, Record{Date: p.Date, PER: p.AdjClose / e, PBR: p.AdjClose / v})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%s: no day had both earnings and book value published", ticker)
	}
	return out, nil
}

// point is a value that became usable on a given date.
type point struct {
	Filed time.Time
	Value float64
}

// trailingEPS builds the TTM earnings series.
//
// Each quarter is restated for splits *individually* before being summed. Using
// one factor for the whole window would corrupt every trailing figure that spans
// a split, since the quarters on either side are quoted in different share
// counts.
func trailingEPS(f Fundamentals, splits []Split) []point {
	quarters := append([]Report(nil), f.QuarterlyEPS...)
	sort.Slice(quarters, func(i, j int) bool { return quarters[i].End.Before(quarters[j].End) })

	var out []point
	for i := 3; i < len(quarters); i++ {
		window := quarters[i-3 : i+1]
		sum, filed := 0.0, window[0].Filed
		for _, q := range window {
			sum += q.Value / splitFactor(splits, q.End)
			if q.Filed.After(filed) {
				filed = q.Filed
			}
		}
		out = append(out, point{Filed: filed, Value: sum})
	}

	// Annual figures cover issuers that never file a standalone fourth quarter.
	for _, a := range f.AnnualEPS {
		out = append(out, point{Filed: a.Filed, Value: a.Value / splitFactor(splits, a.End)})
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Filed.Before(out[j].Filed) })
	return out
}

// bookValuePerShare pairs each equity figure with the share count known at that
// filing, then restates the result on today's split basis.
func bookValuePerShare(f Fundamentals, splits []Split) []point {
	shares := append([]Report(nil), f.Shares...)
	sortReports(shares)

	var out []point
	for _, eq := range f.Equity {
		count := latestBefore(shares, eq.Filed)
		if count <= 0 {
			continue
		}
		out = append(out, point{
			Filed: eq.Filed,
			Value: (eq.Value / count) / splitFactor(splits, eq.Filed),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Filed.Before(out[j].Filed) })
	return out
}

func latestBefore(reports []Report, t time.Time) float64 {
	value := 0.0
	for _, r := range reports {
		if r.Filed.After(t) {
			break
		}
		value = r.Value
	}
	return value
}

// cursor walks a filing-ordered series alongside ascending price dates, so the
// whole series is scanned once instead of once per trading day.
type cursor struct {
	points []point
	i      int
}

func newCursor(points []point) *cursor { return &cursor{points: points} }

func (c *cursor) at(t time.Time) (float64, bool) {
	for c.i < len(c.points) && !c.points[c.i].Filed.After(t) {
		c.i++
	}
	if c.i == 0 {
		return 0, false
	}
	return c.points[c.i-1].Value, true
}
