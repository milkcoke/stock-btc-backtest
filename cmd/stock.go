package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"stock-btc-backtest/backtester"
	"stock-btc-backtest/chart"
	"stock-btc-backtest/downloader"
	"stock-btc-backtest/loader"
	"stock-btc-backtest/reporter"
	"stock-btc-backtest/strategy"
)

const fearGreedCSV = "data/cnn_fear_greed.csv"

var stockCmd = &cobra.Command{
	Use:   "stock",
	Short: "Backtest stock strategies (TQQQ, QLD, QQQ) using CNN Fear & Greed index",
	RunE:  runStock,
}

func runStock(cmd *cobra.Command, args []string) error {
	start, end, err := parseDates()
	if err != nil {
		return err
	}

	years := end.Year() - start.Year()
	lumpSum := float64(years) * 12_000

	tickerConfigs := []struct {
		symbol  string
		path    string
		color   string
		erDelta float64
		dl      func(string) error
	}{
		{"TQQQ", "data/tqqq.csv", "#e74c3c", 0, downloader.TQQQ},
		{"QLD", "data/qld.csv", "#e67e22", 0, downloader.QLD},
		{"QQQ", "data/qqq.csv", "#3498db", 0, downloader.QQQ},
		// KQLD: 2x QQQ Korean ETF, uses QLD price data; TER 0.3372% vs QLD's 0.95% → lower fees
		//{"KQLD", "data/qld.csv", "#2ecc71", 0.003372 - 0.0095, nil},
	}

	rp := reporter.Reporter{}
	var tickerCharts []chart.TickerChart

	for _, t := range tickerConfigs {
		if t.dl != nil {
			if err := downloader.EnsureUpToDate(t.path, t.dl); err != nil {
				log.Fatalf("update %s failed: %v", t.symbol, err)
			}
		}
		prices, indicator, err := loader.StockLoader{
			PricePath:     t.path,
			IndicatorPath: fearGreedCSV,
		}.Load()
		if err != nil {
			log.Fatalf("load %s: %v", t.symbol, err)
		}

		bt := backtester.New(prices, indicator, t.erDelta)
		strats := buildStockStrategies(lumpSum, years)
		results := make([]backtester.Result, len(strats))
		for i, s := range strats {
			results[i] = bt.Run(s, start, end)
		}
		rp.Print(t.symbol, results, start, end)
		tickerCharts = append(tickerCharts, chart.TickerChart{
			Symbol:  t.symbol,
			Results: results,
		})
	}

	const chartPath = "chart_stock.html"
	if err := chart.GenerateStock(chartPath, tickerCharts, start, end); err != nil {
		log.Fatalf("generate chart: %v", err)
	}
	log.Printf("chart saved → %s", chartPath)

	const strategy3Idx = 2
	var compLines []chart.ComparisonLine
	for i, t := range tickerConfigs {
		compLines = append(compLines, chart.ComparisonLine{
			Label:  t.symbol,
			Color:  t.color,
			Result: tickerCharts[i].Results[strategy3Idx],
		})
	}
	strategyName := tickerCharts[0].Results[strategy3Idx].StrategyName
	const compPath = "chart_stock_strategy3.html"
	if err := chart.GenerateStockComparison(compPath, strategyName, compLines, start, end); err != nil {
		log.Fatalf("generate comparison chart: %v", err)
	}
	log.Printf("comparison chart saved → %s", compPath)
	return nil
}

func buildStockStrategies(lumpSum float64, years int) []strategy.Strategy {
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
		&strategy.FearGreedAccumStrategy{
			Label:         "Strategy 4: Save $1,000/month, buy all on F&G <= 24",
			InitialAmount: 0,
			MonthlyAmount: 1_000,
			BuyThreshold:  24,
			SellThreshold: 0,
		},
		&strategy.FearGreedTieredStrategy{
			Label:         "Strategy 5: Save $1,000/month, tiered buy",
			MonthlyAmount: 1_000,
		},
		&strategy.FearGreedAccumStrategy{
			Label:         "Strategy 6: Save $1,000/month, buy all on F&G <= 24, sell on Greed >= 76",
			InitialAmount: 0,
			MonthlyAmount: 1_000,
			BuyThreshold:  24,
			SellThreshold: 76,
		},
		&strategy.FearGreedAccumStrategy{
			Label:         "Strategy 7: Save $1,000/month, buy all on F&G <= 24, sell on Greed >= 60",
			InitialAmount: 0,
			MonthlyAmount: 1_000,
			BuyThreshold:  24,
			SellThreshold: 60,
		},
	}
}
