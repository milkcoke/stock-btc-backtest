package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"stock-btc-backtest/backtester"
	"stock-btc-backtest/chart"
	"stock-btc-backtest/downloader"
	"stock-btc-backtest/loader"
	"stock-btc-backtest/reporter"
	"stock-btc-backtest/strategy"
)

const usdKRWCSV = "data/usd_krw.csv"
const kimchiChartPath = "chart_kimchi.html"
const kimchiChartKoPath = "chart_kimchi_ko.html"
const usdBacktestChartPath = "chart_usd.html"
const usdBacktestKoChartPath = "chart_usd_ko.html"

var usdCmd = &cobra.Command{
	Use:   "usd",
	Short: "Download USD/KRW and USDT/KRW exchange rates, generate kimchi premium chart and backtest",
	RunE:  runUSD,
}

func init() {
	rootCmd.AddCommand(usdCmd)
}

func runUSD(cmd *cobra.Command, args []string) error {
	start, end, err := parseDates()
	if err != nil {
		return err
	}

	log.Printf("Downloading USD/KRW and USDT/KRW (%s ~ %s)...", start.Format("2006-01-02"), end.Format("2006-01-02"))
	if err := downloader.USDKRWCombined(usdKRWCSV, start, end); err != nil {
		return err
	}
	log.Printf("Saved → %s", usdKRWCSV)

	records, err := loader.USDLoader{Path: usdKRWCSV}.Load()
	if err != nil {
		return err
	}
	log.Printf("Loaded %d records with USDT data", len(records))

	if err := chart.GenerateKimchi(kimchiChartPath, records); err != nil {
		return err
	}
	log.Printf("Chart saved → %s", kimchiChartPath)
	if err := chart.GenerateKimchiKorean(kimchiChartKoPath, records); err != nil {
		return err
	}
	log.Printf("Chart saved → %s", kimchiChartKoPath)

	// price = USDT/KRW (actual transaction price on Upbit)
	// pctMap  = PremiumPct  (%) indicator  → strategies 1-8
	// wonMap  = PremiumWon  (₩) indicator  → strategies 9-14
	// rateMap = USD/KRW rate  indicator    → strategies 15-17 (trigger on USD rate, transact at USDT price)
	prices, pctMap, wonMap, rateMap := convertKimchiToBacktest(records)

	btPct := backtester.New(prices, pctMap, 0).WithTradingFee(0.0005)
	btWon := backtester.New(prices, wonMap, 0).WithTradingFee(0.0005)
	btRate := backtester.New(prices, rateMap, 0).WithTradingFee(0.0005)

	months := monthsBetween(start, end)
	pctStrats, wonStrats, rateStrats := buildUSDStrategies(months)

	var results []backtester.Result
	for _, s := range pctStrats {
		results = append(results, btPct.Run(s, start, end))
	}
	for _, s := range wonStrats {
		results = append(results, btWon.Run(s, start, end))
	}
	for _, s := range rateStrats {
		results = append(results, btRate.Run(s, start, end))
	}

	// Contributions and portfolio values here are won amounts, not dollars.
	reporter.Reporter{Currency: "₩", CurrencyLabel: "KRW"}.Print("USD/KRW", results, start, end)

	if err := chart.GenerateUSD(usdBacktestChartPath, results, start, end); err != nil {
		return err
	}
	log.Printf("Chart saved → %s", usdBacktestChartPath)
	if err := chart.GenerateUSDKorean(usdBacktestKoChartPath, results, start, end); err != nil {
		return err
	}
	log.Printf("Chart saved → %s", usdBacktestKoChartPath)
	return nil
}

func convertKimchiToBacktest(records []loader.KimchiRecord) (
	prices []loader.PriceRecord,
	pctMap, wonMap, rateMap map[string]float64,
) {
	prices = make([]loader.PriceRecord, len(records))
	pctMap = make(map[string]float64, len(records))
	wonMap = make(map[string]float64, len(records))
	rateMap = make(map[string]float64, len(records))
	for i, r := range records {
		t, _ := time.Parse("2006-01-02", r.Date)
		prices[i] = loader.PriceRecord{Date: t, AdjClose: r.USDTKRW} // transaction price = USDT/KRW
		pctMap[r.Date] = r.PremiumPct
		wonMap[r.Date] = r.PremiumWon
		rateMap[r.Date] = r.USDKRW // USD/KRW rate used as trigger for strategies 15-17
	}
	return
}

func monthsBetween(start, end time.Time) int {
	m := (end.Year()-start.Year())*12 + int(end.Month()-start.Month())
	if m <= 0 {
		return 1
	}
	return m
}

// buildUSDStrategies returns three slices in README order:
// pctStrats (1-8), wonStrats (9-14), rateStrats (15-17).
func buildUSDStrategies(months int) (pctStrats, wonStrats, rateStrats []strategy.Strategy) {
	const initial = 10_000_000.0
	monthlyDCA := initial / float64(months)

	pctStrats = []strategy.Strategy{
		// 1: buy USDT on day 1 with 5M KRW, hold forever
		// final value = last day USDT/KRW price × USDT count
		&strategy.LumpSumStrategy{
			Label:  "Strategy 1: Buy ₩10,000,000 of USDT on day 1, hold forever",
			Amount: initial,
		},
		// 2: spread ₩10M equally over the backtest period, DCA on 25th monthly
		&strategy.MonthlyDCAStrategy{
			Label:         fmt.Sprintf("Strategy 2: ₩10,000,000/%d months = ₩%.0f/month on 25th", months, monthlyDCA),
			InitialAmount: 0,
			MonthlyAmount: monthlyDCA,
			DayOfMonth:    25,
		},
		// 3-8: start with ₩5M; buy all cash at USDT/KRW price on kimchi premium % trigger
		&strategy.KimchiStrategy{Label: "Strategy 3: ₩10M initial, buy on Premium <= 0%, sell >= 2%", InitialAmount: initial, BuyThreshold: 0, SellThreshold: 2},
		&strategy.KimchiStrategy{Label: "Strategy 4: ₩10M initial, buy on Premium <= -1%, sell >= 2%", InitialAmount: initial, BuyThreshold: -1, SellThreshold: 2},
		&strategy.KimchiStrategy{Label: "Strategy 5: ₩10M initial, buy on Premium <= -2%, sell >= 2%", InitialAmount: initial, BuyThreshold: -2, SellThreshold: 2},
		&strategy.KimchiStrategy{Label: "Strategy 6: ₩10M initial, buy on Premium <= 0%, sell >= 3%", InitialAmount: initial, BuyThreshold: 0, SellThreshold: 3},
		&strategy.KimchiStrategy{Label: "Strategy 7: ₩10M initial, buy on Premium <= -1%, sell >= 3%", InitialAmount: initial, BuyThreshold: -1, SellThreshold: 3},
		&strategy.KimchiStrategy{Label: "Strategy 8: ₩10M initial, buy on Premium <= -2%, sell >= 3%", InitialAmount: initial, BuyThreshold: -2, SellThreshold: 3},
	}

	wonStrats = []strategy.Strategy{
		&strategy.KimchiStrategy{Label: "Strategy 9: ₩10M initial, buy on Premium <= ₩10, sell >= ₩21", InitialAmount: initial, BuyThreshold: 10, SellThreshold: 21},
		&strategy.KimchiStrategy{Label: "Strategy 10: ₩10M initial, buy on Premium <= ₩10, sell >= ₩30", InitialAmount: initial, BuyThreshold: 10, SellThreshold: 30},
		&strategy.KimchiStrategy{Label: "Strategy 11: ₩10M initial, buy on Premium <= ₩10, sell >= ₩40", InitialAmount: initial, BuyThreshold: 10, SellThreshold: 40},
		&strategy.KimchiStrategy{Label: "Strategy 12: ₩10M initial, buy on Premium <= ₩5, sell >= ₩21", InitialAmount: initial, BuyThreshold: 5, SellThreshold: 21},
		&strategy.KimchiStrategy{Label: "Strategy 13: ₩10M initial, buy on Premium <= ₩4, sell >= ₩21", InitialAmount: initial, BuyThreshold: 4, SellThreshold: 21},
		&strategy.KimchiStrategy{Label: "Strategy 14: ₩10M initial, buy on Premium <= ₩3, sell >= ₩21", InitialAmount: initial, BuyThreshold: 3, SellThreshold: 21},
		&strategy.KimchiStrategy{Label: "Strategy 15: ₩10M initial, buy on Premium <= ₩2, sell >= ₩21", InitialAmount: initial, BuyThreshold: 2, SellThreshold: 21},
		&strategy.KimchiStrategy{Label: "Strategy 16: ₩10M initial, buy on Premium <= ₩1, sell >= ₩21", InitialAmount: initial, BuyThreshold: 1, SellThreshold: 21},
		&strategy.KimchiStrategy{Label: "Strategy 17: ₩10M initial, buy on Premium <= ₩0, sell >= ₩21", InitialAmount: initial, BuyThreshold: 0, SellThreshold: 21},
		&strategy.KimchiStrategy{Label: "Strategy 18: ₩10M initial, buy on Premium <= ₩5, sell >= ₩30", InitialAmount: initial, BuyThreshold: 5, SellThreshold: 30},
		&strategy.KimchiStrategy{Label: "Strategy 19: ₩10M initial, buy on Premium <= ₩5, sell >= ₩40", InitialAmount: initial, BuyThreshold: 5, SellThreshold: 40},
		&strategy.KimchiStrategy{Label: "Strategy 20: ₩10M initial, buy on Premium <= -₩10, sell >= ₩10", InitialAmount: initial, BuyThreshold: -10, SellThreshold: 10},
		&strategy.KimchiStrategy{Label: "Strategy 21: ₩10M initial, buy on Premium <= -₩10, sell >= ₩21", InitialAmount: initial, BuyThreshold: -10, SellThreshold: 21},
		&strategy.KimchiStrategy{Label: "Strategy 22: ₩10M initial, buy on Premium <= -₩10, sell >= ₩30", InitialAmount: initial, BuyThreshold: -10, SellThreshold: 30},
	}

	rateStrats = []strategy.Strategy{
		&strategy.USDKRWStrategy{Label: "Strategy 23: ₩10M initial, buy on USD/KRW <= ₩1400, sell >= ₩1500", InitialAmount: initial, BuyRate: 1400, SellRate: 1500},
		&strategy.USDKRWStrategy{Label: "Strategy 24: ₩10M initial, buy on USD/KRW <= ₩1450, sell >= ₩1500", InitialAmount: initial, BuyRate: 1450, SellRate: 1500},
		&strategy.USDKRWStrategy{Label: "Strategy 25: ₩10M initial, buy on USD/KRW <= ₩1420, sell >= ₩1475", InitialAmount: initial, BuyRate: 1420, SellRate: 1475},
	}
	return
}
