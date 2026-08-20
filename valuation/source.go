package valuation

import (
	"sort"
	"time"
)

// Report is one reported figure, tagged with the date it became public.
//
// Filed is what makes the derived series honest: a ratio computed from numbers
// nobody had yet is a look into the future, and an entry rule tested against it
// would look far better than it could ever have been.
type Report struct {
	End   time.Time // period end
	Filed time.Time // date the filing became public
	Value float64

	periodDays int // duration of the reported period; distinguishes a quarter from a year
}

// Fundamentals is everything needed to derive PER and PBR.
type Fundamentals struct {
	QuarterlyEPS []Report // diluted EPS, ~3-month periods, as reported
	AnnualEPS    []Report // diluted EPS, full year — covers issuers that never file a Q4
	Equity       []Report // stockholders' equity
	Shares       []Report // shares outstanding at filing time
}

// Split is a share split. Ratio is the share-count multiplier (3-for-1 → 3).
type Split struct {
	Date  time.Time
	Ratio float64
}

// FundamentalsSource fetches reported figures for a ticker.
type FundamentalsSource interface {
	Fetch(ticker string) (Fundamentals, error)
}

// SplitSource fetches the split history for a ticker.
type SplitSource interface {
	Splits(ticker string) ([]Split, error)
}

// splitFactor is the cumulative share multiplier from splits *after* t.
//
// Per-share figures reported before a split are stated in pre-split shares, so
// dividing by this factor restates them on today's basis — the same basis the
// back-adjusted price series already uses.
func splitFactor(splits []Split, t time.Time) float64 {
	f := 1.0
	for _, s := range splits {
		if s.Date.After(t) {
			f *= s.Ratio
		}
	}
	return f
}

// sortReports orders by filing date, then period end, which is the order the
// market learned the figures in.
func sortReports(reports []Report) {
	sort.Slice(reports, func(i, j int) bool {
		if reports[i].Filed.Equal(reports[j].Filed) {
			return reports[i].End.Before(reports[j].End)
		}
		return reports[i].Filed.Before(reports[j].Filed)
	})
}
