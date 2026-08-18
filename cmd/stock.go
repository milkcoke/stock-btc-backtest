package cmd

import (
	"fmt"
	"log"

	"stock-btc-backtest/backtester"
	"stock-btc-backtest/chart"
	"stock-btc-backtest/downloader"
	"stock-btc-backtest/loader"
	"stock-btc-backtest/reporter"
	"stock-btc-backtest/strategy"

	"github.com/spf13/cobra"
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

	// Each ticker is backtested in the currency it is quoted in — Korean listings
	// stay in KRW and are never converted to USD. monthly is the contribution in
	// that same currency, sized so the strategies are comparable in real terms.
	tickerConfigs := []struct {
		symbol        string
		path          string
		color         string
		erDelta       float64
		currency      string
		currencyLabel string
		monthly       float64
		dl            func(string) error
	}{
		{"TQQQ", "data/tqqq.csv", "#e74c3c", 0, "$", "USD", 1_000, downloader.TQQQ},
		{"QLD", "data/qld.csv", "#e67e22", 0, "$", "USD", 1_000, downloader.QLD},
		{"QQQ", "data/qqq.csv", "#3498db", 0, "$", "USD", 1_000, downloader.QQQ},
		// KIH: 한국금융지주 (Korea Investment Holdings), KOSPI 071050 — quoted in KRW.
		{"KIH", "data/kih.csv", "#9b59b6", 0, "₩", "KRW", 1_000_000, downloader.KIH},
		// KQLD: 2x QQQ Korean ETF, uses QLD price data; TER 0.3372% vs QLD's 0.95% → lower fees
		//{"KQLD", "data/qld.csv", "#2ecc71", 0.003372 - 0.0095, "$", "USD", 1_000, nil},
	}

	if err := downloader.EnsureUpToDate(fearGreedCSV, downloader.FearGreed); err != nil {
		log.Fatalf("update CNN Fear & Greed failed: %v", err)
	}

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
		strats := buildStockStrategies(t.currency, t.monthly, years)
		results := make([]backtester.Result, len(strats))
		for i, s := range strats {
			results[i] = bt.Run(s, start, end)
		}
		reporter.Reporter{Currency: t.currency, CurrencyLabel: t.currencyLabel}.
			Print(t.symbol, results, start, end)
		tickerCharts = append(tickerCharts, chart.TickerChart{
			Symbol:        t.symbol,
			Results:       results,
			Currency:      t.currency,
			CurrencyLabel: t.currencyLabel,
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
			Label:    t.symbol,
			Color:    t.color,
			Currency: t.currency,
			Result:   tickerCharts[i].Results[strategy3Idx],
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

// buildStockStrategies builds the strategy set for one ticker. cur is the
// currency symbol the ticker is quoted in and monthly the contribution in that
// currency; every amount below is derived from monthly so the same strategies
// run at USD scale for US listings and KRW scale for Korean ones.
func buildStockStrategies(cur string, monthly float64, years int) []strategy.Strategy {
	annual := monthly * 12
	lumpSum := annual * float64(years)
	m := money(cur, monthly)

	return []strategy.Strategy{
		&strategy.LumpSumStrategy{
			Label: fmt.Sprintf("Strategy 1: Lump-sum %s (%d yrs × %s) on day 1",
				money(cur, lumpSum), years, money(cur, annual)),
			Amount: lumpSum,
		},
		&strategy.AnnualDCAStrategy{
			Label:         fmt.Sprintf("Strategy 2: %s every Jan 1st", money(cur, annual)),
			InitialAmount: 0,
			AnnualAmount:  annual,
		},
		&strategy.MonthlyDCAStrategy{
			Label:         fmt.Sprintf("Strategy 3: %s on 25th monthly", m),
			InitialAmount: 0,
			MonthlyAmount: monthly,
			DayOfMonth:    25,
		},
		&strategy.FearGreedAccumStrategy{
			Label:         fmt.Sprintf("Strategy 4: Save %s/month, buy all on F&G <= 24", m),
			InitialAmount: 0,
			MonthlyAmount: monthly,
			BuyThreshold:  24,
			SellThreshold: 0,
		},
		&strategy.FearGreedTieredStrategy{
			Label:         fmt.Sprintf("Strategy 5: Save %s/month, tiered buy", m),
			MonthlyAmount: monthly,
		},
		&strategy.FearGreedAccumStrategy{
			Label:         fmt.Sprintf("Strategy 6: Save %s/month, buy all on F&G <= 24, sell on Greed >= 76", m),
			InitialAmount: 0,
			MonthlyAmount: monthly,
			BuyThreshold:  24,
			SellThreshold: 76,
		},
		&strategy.FearGreedAccumStrategy{
			Label:         fmt.Sprintf("Strategy 7: Save %s/month, buy all on F&G <= 24, sell on Greed >= 60", m),
			InitialAmount: 0,
			MonthlyAmount: monthly,
			BuyThreshold:  24,
			SellThreshold: 60,
		},
	}
}

// money renders a whole amount with thousands separators, e.g. "₩1,000,000".
func money(cur string, v float64) string {
	s := fmt.Sprintf("%.0f", v)
	out := make([]byte, 0, len(s)+len(s)/3)
	for i := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, s[i])
	}
	return cur + string(out)
}
