package loader

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
)

type KimchiRecord struct {
	Date       string
	USDKRW     float64
	USDTKRW    float64
	PremiumWon float64 // USDTKRW - USDKRW
	PremiumPct float64 // (USDTKRW - USDKRW) / USDKRW * 100
}

type USDLoader struct {
	Path string
}

func (l USDLoader) Load() ([]KimchiRecord, error) {
	f, err := os.Open(l.Path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, err
	}

	var records []KimchiRecord
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) < 3 || row[2] == "" {
			continue // no Upbit data for this date
		}
		usd, err := strconv.ParseFloat(row[1], 64)
		if err != nil {
			return nil, fmt.Errorf("row %d usd_krw: %w", i, err)
		}
		usdt, err := strconv.ParseFloat(row[2], 64)
		if err != nil {
			return nil, fmt.Errorf("row %d usdt_krw: %w", i, err)
		}
		premWon := usdt - usd
		records = append(records, KimchiRecord{
			Date:       row[0],
			USDKRW:     usd,
			USDTKRW:    usdt,
			PremiumWon: premWon,
			PremiumPct: premWon / usd * 100,
		})
	}
	return records, nil
}
