package chart

import (
	"time"

	"stock-btc-backtest/backtester"
)

func GenerateBTC(outputPath string, results []backtester.Result, start, end time.Time) error {
	return Generate(outputPath, []TickerChart{{Symbol: "BTC", Results: results}}, start, end)
}
