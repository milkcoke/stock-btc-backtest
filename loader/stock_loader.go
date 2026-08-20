package loader

import (
	"encoding/csv"
	"os"
	"strconv"
)

// StockLoader loads price data and the Fear & Greed indicator from separate CSVs.
type StockLoader struct {
	PricePath     string
	IndicatorPath string
}

func (l StockLoader) Load() ([]PriceRecord, map[string]float64, error) {
	prices, err := loadPrices(l.PricePath)
	if err != nil {
		return nil, nil, err
	}
	fearGreed, err := loadFearGreed(l.IndicatorPath)
	if err != nil {
		return nil, nil, err
	}
	return prices, fearGreed, nil
}

func loadPrices(path string) ([]PriceRecord, error) {
	return readPriceCSV(path, "AdjClose")
}

func loadFearGreed(path string) (map[string]float64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, err
	}

	m := make(map[string]float64, len(rows)-1)
	for _, row := range rows[1:] {
		val, err := strconv.ParseFloat(row[1], 64)
		if err != nil {
			continue
		}
		m[row[0]] = val
	}
	return m, nil
}
