package chart

import (
	"time"

	"stock-btc-backtest/backtester"
)

func GenerateUSD(outputPath string, results []backtester.Result, start, end time.Time) error {
	return Generate(outputPath, []TickerChart{{
		Symbol:        "USD-KRW",
		Results:       results,
		Currency:      "₩",
		CurrencyLabel: "KRW",
	}}, start, end)
}
