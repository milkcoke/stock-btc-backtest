package eventstudy

import (
	"fmt"
	"time"

	"stock-btc-backtest/indicator"
	"stock-btc-backtest/loader"
	"stock-btc-backtest/signal"
	"stock-btc-backtest/stats"
)

// Study configures one run. Prices must be the *full* history, not the window:
// the moving average and RSI need history in front of the window to be warm on
// its first day.
type Study struct {
	Symbol      string
	PriceColumn string
	Prices      []loader.PriceRecord

	Years         int // preferred window, e.g. 20
	FallbackYears int // used when the history is shorter, e.g. 10

	MAWindow   int
	MADiscount float64 // 0.10 = 10% below the average
	RSIPeriod  int
	RSIMax     float64

	// MDDThreshold overrides the derived threshold when > 0. Zero means "use the
	// average of this window's yearly drawdowns".
	MDDThreshold float64

	Targets []float64
}

// EntryResult is one entry plus what happened to it.
type EntryResult struct {
	signal.Entry
	Outcomes []Outcome
	LowDate  time.Time
	Low      float64
	LowPct   float64
}

// TargetStat aggregates one target across all entries. Achieved and Total are
// kept apart from Summary.N so the report can say how many entries were left
// out rather than just showing a smaller sample.
type TargetStat struct {
	Target   float64
	Summary  stats.Summary
	Achieved int
	Total    int
}

// Result carries everything the report and chart need and nothing about how to
// display it.
type Result struct {
	Symbol      string
	PriceColumn string
	From, To    time.Time
	Years       int
	Rule        string

	MDDThreshold float64 // fraction actually used
	MDDDerived   bool    // true when it came from the data rather than a flag
	YearlyMDD    []indicator.YearMDD
	MDDSummary   stats.Summary // complete calendar years only

	Entries []EntryResult
	Targets []TargetStat

	// Windowed series, for charting.
	Dates  []time.Time
	Prices []float64
	SMA    []float64
}

// Run executes the study. The order below is the whole point of this type:
// indicators first (over all history), then the window, then the threshold that
// the window implies, and only then the rule that depends on it.
func (s Study) Run() (Result, error) {
	if len(s.Prices) == 0 {
		return Result{}, fmt.Errorf("%s: no price data", s.Symbol)
	}

	dates := make([]time.Time, len(s.Prices))
	prices := make([]float64, len(s.Prices))
	for i, p := range s.Prices {
		dates[i], prices[i] = p.Date, p.AdjClose
	}

	sma := indicator.SMA(prices, s.MAWindow)
	rsi := indicator.WilderRSI(prices, s.RSIPeriod)
	peak := indicator.RunningPeak(dates, prices)

	win, err := ResolveWindow(dates, s.Years, s.FallbackYears)
	if err != nil {
		return Result{}, fmt.Errorf("%s: %w", s.Symbol, err)
	}

	yearly := indicator.YearlyMDD(dates[win.From:win.To+1], prices[win.From:win.To+1])
	mddSummary := summarizeComplete(yearly)

	threshold, derived := s.MDDThreshold, false
	if threshold <= 0 {
		if mddSummary.N == 0 {
			return Result{}, fmt.Errorf("%s: no complete calendar year in the window to derive a threshold from", s.Symbol)
		}
		threshold, derived = -mddSummary.Mean/100, true
	}

	rule := signal.AllOf{
		signal.DrawdownAtLeast{Pct: threshold},
		signal.MADiscount{Pct: s.MADiscount, Window: s.MAWindow},
		signal.RSIBelow{Max: s.RSIMax, Period: s.RSIPeriod},
	}
	frame := signal.Frame{Dates: dates, Prices: prices, SMA: sma, RSI: rsi, Peak: peak}
	entries := signal.Detector{Condition: rule}.Detect(frame, win.From, win.To)

	tracker := TargetTracker{Targets: s.Targets}
	results := make([]EntryResult, 0, len(entries))
	for _, e := range entries {
		outcomes := tracker.Track(prices, dates, e.Index, win.To)
		lowDate, low := LowestUntil(prices, dates, e.Index, holdingEndIndex(outcomes, win.To))
		results = append(results, EntryResult{
			Entry: e, Outcomes: outcomes,
			LowDate: lowDate, Low: low, LowPct: (low/e.Buy - 1) * 100,
		})
	}

	return Result{
		Symbol: s.Symbol, PriceColumn: s.PriceColumn,
		From: dates[win.From], To: dates[win.To], Years: win.Years,
		Rule:         rule.Describe(),
		MDDThreshold: threshold, MDDDerived: derived,
		YearlyMDD: yearly, MDDSummary: mddSummary,
		Entries: results, Targets: aggregate(s.Targets, results),
		Dates: dates[win.From : win.To+1], Prices: prices[win.From : win.To+1],
		SMA: sma[win.From : win.To+1],
	}, nil
}

// holdingEndIndex is the last day the position was still working towards a
// target in the report.
//
// Bounding at the *first* target instead would make the drawdown column read
// 0% for any entry that bounced quickly, even when the same row shows a target
// that took years — the position clearly did sit through something. Taking the
// furthest target actually reached covers the whole ride the table describes.
func holdingEndIndex(outcomes []Outcome, fallback int) int {
	end := fallback
	for i := len(outcomes) - 1; i >= 0; i-- {
		if outcomes[i].Achieved {
			return outcomes[i].HitIndex
		}
	}
	return end
}

func summarizeComplete(yearly []indicator.YearMDD) stats.Summary {
	var vals []float64
	for _, y := range yearly {
		if y.Complete {
			vals = append(vals, y.Pct)
		}
	}
	return stats.Summarize(vals)
}

func aggregate(targets []float64, results []EntryResult) []TargetStat {
	out := make([]TargetStat, 0, len(targets))
	for i, target := range targets {
		var days []float64
		achieved := 0
		for _, r := range results {
			if i < len(r.Outcomes) && r.Outcomes[i].Achieved {
				days = append(days, float64(r.Outcomes[i].Days))
				achieved++
			}
		}
		out = append(out, TargetStat{Target: target, Summary: stats.Summarize(days),
			Achieved: achieved, Total: len(results)})
	}
	return out
}
