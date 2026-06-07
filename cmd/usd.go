package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"stock-btc-backtest/chart"
	"stock-btc-backtest/downloader"
	"stock-btc-backtest/loader"
)

const usdKRWCSV = "data/usd_krw.csv"
const kimchiChartPath = "chart_kimchi.html"

var usdCmd = &cobra.Command{
	Use:   "usd",
	Short: "Download USD/KRW and USDT/KRW exchange rates and generate kimchi premium chart",
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
	return nil
}
