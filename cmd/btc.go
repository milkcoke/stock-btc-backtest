package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"stock-btc-backtest/backtester"
	"stock-btc-backtest/chart"
	"stock-btc-backtest/loader"
	"stock-btc-backtest/reporter"
	"stock-btc-backtest/strategy"
)

const btcMVRVCSV = "data/MVRV_Z-Score.csv"

var btcCmd = &cobra.Command{
	Use:   "btc",
	Short: "Backtest BTC strategies using MVRV Z-Score",
	RunE:  runBTC,
}

func runBTC(cmd *cobra.Command, args []string) error {
	start, end, err := parseDates()
	if err != nil {
		return err
	}

	prices, mvrv, err := loader.BTCLoader{Path: btcMVRVCSV}.Load()
	if err != nil {
		log.Fatalf("load BTC data: %v", err)
	}

	years := end.Year() - start.Year()
	lumpSum := float64(years) * 12_000

	bt := backtester.New(prices, mvrv, 0)
	strats := buildBTCStrategies(lumpSum, years)
	results := make([]backtester.Result, len(strats))
	for i, s := range strats {
		results[i] = bt.Run(s, start, end)
	}

	rp := reporter.Reporter{}
	rp.Print("BTC", results, start, end)

	const chartPath = "chart_btc.html"
	if err := chart.Generate(chartPath, []chart.TickerChart{{Symbol: "BTC", Results: results}}, start, end); err != nil {
		log.Fatalf("generate chart: %v", err)
	}
	log.Printf("chart saved → %s", chartPath)
	return nil
}

func buildBTCStrategies(lumpSum float64, years int) []strategy.Strategy {
	return []strategy.Strategy{
		&strategy.LumpSumStrategy{
			Label:  fmt.Sprintf("Strategy 1: Lump-sum $%.0f (%d yrs × $12,000) on day 1", lumpSum, years),
			Amount: lumpSum,
		},
		&strategy.AnnualDCAStrategy{
			InitialAmount: 0,
			AnnualAmount:  12_000,
		},
		&strategy.MonthlyDCAStrategy{
			Label:         "Strategy 3: $1,000 on 25th monthly",
			InitialAmount: 0,
			MonthlyAmount: 1_000,
			DayOfMonth:    25,
		},
		&strategy.MVRVAccumStrategy{
			Label:         "Strategy 4: Save $1,000/month, buy all on MVRV <= 0, never sell",
			MonthlyAmount: 1_000,
			BuyThreshold:  0,
			SellThreshold: 0,
		},
		&strategy.MVRVAccumStrategy{
			Label:         "Strategy 5: Save $1,000/month, buy all on MVRV <= 0, sell on >= 7",
			MonthlyAmount: 1_000,
			BuyThreshold:  0,
			SellThreshold: 7,
		},
		&strategy.MVRVAccumStrategy{
			Label:         "Strategy 6: Save $1,000/month, buy all on MVRV <= 0, sell on >= 6",
			MonthlyAmount: 1_000,
			BuyThreshold:  0,
			SellThreshold: 6,
		},
		&strategy.MVRVAccumStrategy{
			Label:         "Strategy 7: Save $1,000/month, buy all on MVRV <= 0, sell on >= 3.5",
			MonthlyAmount: 1_000,
			BuyThreshold:  0,
			SellThreshold: 3.5,
		},
		&strategy.MVRVAccumStrategy{
			Label:         "Strategy 9: Save $1,000/month, buy all on MVRV <= 0.5, never sell",
			MonthlyAmount: 1_000,
			BuyThreshold:  0.5,
			SellThreshold: 0,
		},
		&strategy.MVRVAccumStrategy{
			Label:         "Strategy 10: Save $1,000/month, buy all on MVRV <= 0.5, sell on >= 7",
			MonthlyAmount: 1_000,
			BuyThreshold:  0.5,
			SellThreshold: 7,
		},
		&strategy.MVRVAccumStrategy{
			Label:         "Strategy 11: Save $1,000/month, buy all on MVRV <= 0.5, sell on >= 6",
			MonthlyAmount: 1_000,
			BuyThreshold:  0.5,
			SellThreshold: 6,
		},
		&strategy.MVRVAccumStrategy{
			Label:         "Strategy 12: Save $1,000/month, buy all on MVRV <= 0.5, sell on >= 3.5",
			MonthlyAmount: 1_000,
			BuyThreshold:  0.5,
			SellThreshold: 3.5,
		},
		&strategy.MVRVAccumStrategy{
			Label:         "Strategy 14: Save $1,000/month, buy all on MVRV <= 1, never sell",
			MonthlyAmount: 1_000,
			BuyThreshold:  1,
			SellThreshold: 0,
		},
		&strategy.MVRVAccumStrategy{
			Label:         "Strategy 15: Save $1,000/month, buy all on MVRV <= 1, sell on >= 7",
			MonthlyAmount: 1_000,
			BuyThreshold:  1,
			SellThreshold: 7,
		},
		&strategy.MVRVAccumStrategy{
			Label:         "Strategy 16: Save $1,000/month, buy all on MVRV <= 1, sell on >= 6",
			MonthlyAmount: 1_000,
			BuyThreshold:  1,
			SellThreshold: 6,
		},
		&strategy.MVRVAccumStrategy{
			Label:         "Strategy 17: Save $1,000/month, buy all on MVRV <= 1, sell on >= 3.5",
			MonthlyAmount: 1_000,
			BuyThreshold:  1,
			SellThreshold: 3.5,
		},
	}
}
