package reporter

import (
	"fmt"
	"stock-btc-backtest/backtester"
	"strings"
	"time"
)

type Reporter struct{}

func (Reporter) Print(symbol string, results []backtester.Result, start, end time.Time) {
	fmt.Printf("\n[ %s ]  %s → %s\n\n", symbol, start.Format("2006-01-02"), end.Format("2006-01-02"))

	const col0, col1, col2, col3, col4, col5 = 56, 18, 20, 13, 10, 10
	header := fmt.Sprintf("%-*s %*s %*s %*s %*s %*s", col0, "Strategy", col1, "Invested (USD)", col2, "Final Value (USD)", col3, "Return (%)", col4, "MDD (%)", col5, "MDD Date")
	fmt.Println(header)
	fmt.Println(strings.Repeat("-", len(header)))

	for _, r := range results {
		fmt.Printf("%-*s %*s %*s %*.2f%% %*.2f%% %*s\n",
			col0, r.StrategyName,
			col1, fmt.Sprintf("$%.2f", r.TotalInvested),
			col2, fmt.Sprintf("$%.2f", r.FinalValue),
			col3-1, r.ReturnPct(),
			col4-1, r.MDD,
			col5, r.MDDDate,
		)
	}
	fmt.Println()
}
