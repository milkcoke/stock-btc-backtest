package loader

import (
	"encoding/csv"
	"os"
	"strconv"
	"time"
)

// BTCLoader loads BTC price and MVRV Z-Score from a single CSV
// (columns: Date, Price (USD), MVRV Z-Score).
type BTCLoader struct {
	Path string
}

func (l BTCLoader) Load() ([]PriceRecord, map[string]float64, error) {
	f, err := os.Open(l.Path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, nil, err
	}

	records := make([]PriceRecord, 0, len(rows)-1)
	mvrv := make(map[string]float64, len(rows)-1)
	for _, row := range rows[1:] {
		if len(row) < 2 {
			continue
		}
		date, err := time.Parse("2006-01-02", row[0])
		if err != nil {
			continue
		}
		price, err := strconv.ParseFloat(row[1], 64)
		if err != nil {
			continue
		}
		records = append(records, PriceRecord{Date: date, AdjClose: price})
		if len(row) > 2 {
			if z, err := strconv.ParseFloat(row[2], 64); err == nil {
				mvrv[row[0]] = z
			}
		}
	}
	return records, mvrv, nil
}
