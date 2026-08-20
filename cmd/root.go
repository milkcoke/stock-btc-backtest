package cmd

import (
	"time"

	"github.com/spf13/cobra"
)

var (
	startStr string
	endStr   string
)

var rootCmd = &cobra.Command{
	Use:   "stock-btc-backtest",
	Short: "Backtest investment strategies for stocks and BTC",
	RunE:  runStock,
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.PersistentFlags().StringVar(&startStr, "start", "2011-01-03", "backtest start date (YYYY-MM-DD)")
	rootCmd.PersistentFlags().StringVar(&endStr, "end", time.Now().Format("2006-01-02"), "backtest end date (YYYY-MM-DD)")
	rootCmd.AddCommand(stockCmd)
	rootCmd.AddCommand(btcCmd)
	rootCmd.AddCommand(entryCmd)
}

func parseDates() (start, end time.Time, err error) {
	start, err = time.Parse("2006-01-02", startStr)
	if err != nil {
		return
	}
	end, err = time.Parse("2006-01-02", endStr)
	return
}
