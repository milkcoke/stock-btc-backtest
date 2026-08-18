package reporter

import (
	"fmt"
	"stock-btc-backtest/backtester"
	"strings"
	"time"
	"unicode/utf8"
)

// Reporter prints backtest results. The zero value reports in USD.
type Reporter struct {
	Currency      string // symbol prefix, e.g. "$" or "₩" (default "$")
	CurrencyLabel string // column header label, e.g. "USD" or "KRW" (default "USD")
}

func (r Reporter) Print(symbol string, results []backtester.Result, start, end time.Time) {
	cur, curLabel := r.Currency, r.CurrencyLabel
	if cur == "" {
		cur = "$"
	}
	if curLabel == "" {
		curLabel = "USD"
	}

	fmt.Printf("\n[ %s ]  %s → %s\n\n", symbol, start.Format("2006-01-02"), end.Format("2006-01-02"))

	const col0, col1, col2, col3, col4, col5 = 56, 18, 20, 13, 10, 10
	header := fmt.Sprintf("%-*s %*s %*s %*s %*s %*s",
		col0, "Strategy",
		col1, "Invested ("+curLabel+")",
		col2, "Final Value ("+curLabel+")",
		col3, "Return (%)", col4, "MDD (%)", col5, "MDD Date")
	fmt.Println(header)
	fmt.Println(strings.Repeat("-", utf8.RuneCountInString(header)))

	for _, res := range results {
		fmt.Printf("%-*s %s %s %*.2f%% %*.2f%% %*s\n",
			col0, res.StrategyName,
			padLeft(formatMoney(cur, res.TotalInvested), col1),
			padLeft(formatMoney(cur, res.FinalValue), col2),
			col3-1, res.ReturnPct(),
			col4-1, res.MDD,
			col5, res.MDDDate,
		)
	}
	fmt.Println()
}

// formatMoney renders an amount for the given currency symbol. Won amounts are
// whole numbers with thousands separators; other currencies keep two decimals.
func formatMoney(currency string, v float64) string {
	if currency == "₩" {
		return currency + addThousandsSep(fmt.Sprintf("%.0f", v))
	}
	return fmt.Sprintf("%s%.2f", currency, v)
}

func addThousandsSep(s string) string {
	n := len(s)
	out := make([]byte, 0, n+(n-1)/3)
	for i := range s {
		if i > 0 && (n-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, s[i])
	}
	return string(out)
}

// padLeft right-aligns s to width counted in runes, so multi-byte currency
// symbols such as ₩ do not shift the columns.
func padLeft(s string, width int) string {
	if pad := width - utf8.RuneCountInString(s); pad > 0 {
		return strings.Repeat(" ", pad) + s
	}
	return s
}
