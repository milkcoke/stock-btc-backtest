package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"stock-btc-backtest/browser"
	"stock-btc-backtest/chart"
	"stock-btc-backtest/downloader"
	"stock-btc-backtest/eventstudy"
	"stock-btc-backtest/loader"
	"stock-btc-backtest/valuation"

	"github.com/spf13/cobra"
)

var (
	entryTicker       string
	entryYears        int
	entryFallbackYrs  int
	entryMAWindow     int
	entryMADiscount   float64
	entryRSIPeriod    int
	entryRSIMax       float64
	entryTargets      string
	entryPriceCol     string
	entryMDDThreshold float64
	entryOutput       string
	entryNoDerive     bool
)

var entryCmd = &cobra.Command{
	Use:   "entry",
	Short: "Event study: when a ticker got cheap, how long did each profit target take?",
	Long: `Finds every day a ticker met all three entry conditions — down at least the
window's average yearly MDD from its running high, trading below its moving
average by the given discount, and oversold on RSI — then measures how many
calendar days each profit target took from there.

This is not a portfolio simulation. There is no cash flow, no fee and no tax:
it answers "how long did it take", not "how much would I have".`,
	RunE: runEntry,
}

func init() {
	entryCmd.Flags().StringVar(&entryTicker, "ticker", "", "Yahoo symbol, e.g. QQQ or 071050.KS (required)")
	entryCmd.Flags().IntVar(&entryYears, "years", 20, "preferred analysis window in years")
	entryCmd.Flags().IntVar(&entryFallbackYrs, "fallback-years", 10, "window to fall back to when history is shorter")
	entryCmd.Flags().IntVar(&entryMAWindow, "ma-window", 200, "moving average period in trading days")
	entryCmd.Flags().Float64Var(&entryMADiscount, "ma-discount", 0.10, "required discount to the moving average")
	entryCmd.Flags().IntVar(&entryRSIPeriod, "rsi-period", 14, "RSI period")
	entryCmd.Flags().Float64Var(&entryRSIMax, "rsi-max", 35, "maximum RSI at entry")
	entryCmd.Flags().StringVar(&entryTargets, "targets", "10,20,30,40,50", "profit targets in percent, comma separated")
	entryCmd.Flags().StringVar(&entryPriceCol, "price-col", "AdjClose", "price column: AdjClose or Close")
	entryCmd.Flags().Float64Var(&entryMDDThreshold, "mdd-threshold", 0,
		"fixed drawdown threshold as a fraction (0.30 = 30%); 0 derives it from the window's average yearly MDD")
	entryCmd.Flags().StringVar(&entryOutput, "output", "", "chart path (default chart_entry_{ticker}.html)")
	entryCmd.Flags().BoolVar(&entryNoDerive, "no-derive-valuation", false,
		"skip deriving PER/PBR from SEC filings when data/{ticker}_per_pbr.csv is absent")
	_ = entryCmd.MarkFlagRequired("ticker")
}

// runEntry only wires things together. Every decision it could be tempted to
// make — which window, which threshold, what counts as an entry — belongs to
// eventstudy, so that the same analysis is available without a CLI.
func runEntry(_ *cobra.Command, _ []string) error {
	targets, err := parseTargets(entryTargets)
	if err != nil {
		return err
	}

	pricePath, err := downloader.Ticker(entryTicker)
	if err != nil {
		return fmt.Errorf("download %s: %w", entryTicker, err)
	}

	prices, err := loader.PriceCSV{Path: pricePath, Column: entryPriceCol}.LoadPrices()
	if err != nil {
		return err
	}

	result, err := eventstudy.Study{
		Symbol: entryTicker, PriceColumn: entryPriceCol, Prices: prices,
		Years: entryYears, FallbackYears: entryFallbackYrs,
		MAWindow: entryMAWindow, MADiscount: entryMADiscount,
		RSIPeriod: entryRSIPeriod, RSIMax: entryRSIMax,
		MDDThreshold: entryMDDThreshold, Targets: targets,
	}.Run()
	if err != nil {
		return err
	}

	val := loadValuation(entryTicker, result, prices)

	output := entryOutput
	if output == "" {
		output = fmt.Sprintf("chart_entry_%s.html", strings.ReplaceAll(entryTicker, "/", "_"))
	}
	if err := chart.GenerateEntry(output, result, val); err != nil {
		return err
	}
	return browser.Open(output)
}

// loadValuation prefers a local file and derives one from SEC filings when
// there is none.
//
// The local file wins because it may be better data than anything derivable —
// KRX publishes an official daily series for Korean issuers, and no amount of
// arithmetic reproduces it. Deriving is the fallback, and failing to derive is
// normal: non-US listings have no EDGAR presence at all.
func loadValuation(ticker string, res eventstudy.Result, prices []loader.PriceRecord) *valuation.Summary {
	path := valuation.CSVPath(ticker)
	records, err := valuation.CSVLoader{Path: path}.Load()

	if err != nil && os.IsNotExist(err) && !entryNoDerive {
		records, err = deriveValuation(ticker, path, prices)
	}
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "[warn] %s: %v\n", path, err)
		}
		return nil
	}

	summary := valuation.Summarize(records, res.From, res.To, res.Years)
	return &summary
}

func deriveValuation(ticker, path string, prices []loader.PriceRecord) ([]valuation.Record, error) {
	fmt.Fprintf(os.Stderr, "[info] %s: deriving PER/PBR from SEC filings\n", ticker)
	records, err := valuation.Builder{
		Fundamentals: valuation.EDGAR{},
		Splits:       valuation.YahooSplits{},
	}.Build(ticker, prices)
	if err != nil {
		return nil, err
	}
	if err := valuation.SaveCSV(path, records); err != nil {
		fmt.Fprintf(os.Stderr, "[warn] could not cache %s: %v\n", path, err)
	} else {
		fmt.Fprintf(os.Stderr, "[info] cached %d rows to %s\n", len(records), path)
	}
	return records, nil
}

func parseTargets(s string) ([]float64, error) {
	var out []float64
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		v, err := strconv.ParseFloat(part, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid target %q: %w", part, err)
		}
		out = append(out, v)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no targets given")
	}
	return out, nil
}
