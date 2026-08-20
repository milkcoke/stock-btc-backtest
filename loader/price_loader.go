package loader

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"
)

// PriceLoader is the narrow interface for consumers that only need prices.
//
// Loader bundles a sentiment indicator with the prices because the DCA
// backtester always needs both. The event study does not, and forcing it to
// pass an indicator path it will never read would be an interface it does not
// use.
type PriceLoader interface {
	LoadPrices() ([]PriceRecord, error)
}

// PriceCSV reads a downloader-format CSV
// (Date,Open,High,Low,Close,AdjClose,Volume).
//
// Column selects which price lands in PriceRecord.AdjClose. "AdjClose" (the
// default) is dividend-adjusted total return. "Close" is what a price chart
// shows: split-adjusted but not dividend-adjusted, so the numbers match what
// the user remembers paying.
type PriceCSV struct {
	Path   string
	Column string
}

func (p PriceCSV) LoadPrices() ([]PriceRecord, error) {
	column := p.Column
	if column == "" {
		column = "AdjClose"
	}
	return readPriceCSV(p.Path, column)
}

func readPriceCSV(path, column string) ([]PriceRecord, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) < 2 {
		return nil, fmt.Errorf("%s: no price rows", path)
	}

	dateCol, priceCol := columnIndex(rows[0], "Date"), columnIndex(rows[0], column)
	if dateCol < 0 || priceCol < 0 {
		return nil, fmt.Errorf("%s: missing Date or %s column", path, column)
	}

	records := make([]PriceRecord, 0, len(rows)-1)
	for _, row := range rows[1:] {
		if len(row) <= priceCol {
			continue
		}
		date, err := time.Parse("2006-01-02", row[dateCol])
		if err != nil {
			continue
		}
		price, err := strconv.ParseFloat(row[priceCol], 64)
		if err != nil {
			continue
		}
		records = append(records, PriceRecord{Date: date, AdjClose: price})
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("%s: no parseable rows", path)
	}
	return records, nil
}

func columnIndex(header []string, name string) int {
	for i, h := range header {
		if h == name {
			return i
		}
	}
	return -1
}
