package downloader

import (
	"path/filepath"
	"strings"
)

// DataDir is where price CSVs live, matching the paths the existing commands use.
const DataDir = "data"

// CSVPath returns the conventional price file for a ticker: data/{ticker}.csv.
//
// Yahoo suffixes exchange codes with a dot (071050.KS) and the ticker itself can
// carry one (BRK.B), which is fine in a file name but noisy; keep it verbatim so
// the file is recognisable, and only strip separators that would create a
// directory.
func CSVPath(ticker string) string {
	safe := strings.NewReplacer("/", "_", `\`, "_").Replace(ticker)
	return filepath.Join(DataDir, safe+".csv")
}

// Ticker downloads any Yahoo symbol to its conventional CSV path, skipping the
// request when the file is already current. It returns the path so the caller
// can hand it straight to loader.PriceCSV.
func Ticker(symbol string) (string, error) {
	path := CSVPath(symbol)
	return path, EnsureUpToDate(path, func(p string) error { return Stock(symbol, p) })
}
