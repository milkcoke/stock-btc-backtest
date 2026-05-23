package loader

import "time"

type PriceRecord struct {
	Date     time.Time
	AdjClose float64
}

type Loader interface {
	Load() ([]PriceRecord, map[string]float64, error)
}
