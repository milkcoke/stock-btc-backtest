package valuation

import "stock-btc-backtest/downloader"

// YahooSplits adapts the downloader's split feed to SplitSource, so this
// package keeps its own vocabulary and all Yahoo HTTP stays in one place.
type YahooSplits struct{}

func (YahooSplits) Splits(ticker string) ([]Split, error) {
	events, err := downloader.SplitEvents(ticker)
	if err != nil {
		return nil, err
	}
	out := make([]Split, 0, len(events))
	for _, e := range events {
		out = append(out, Split{Date: e.Date, Ratio: e.Ratio})
	}
	return out, nil
}
