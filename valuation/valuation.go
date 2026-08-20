// Package valuation loads a PER/PBR series and summarises it.
//
// Acquisition is deliberately out of scope. KRX now requires a logged-in
// session for its daily PER/PBR series and the free alternatives only reach
// back four to six years, so the file is produced by hand and this package
// simply reads whatever is there — and reports how much of the requested window
// it actually covers.
package valuation

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"stock-btc-backtest/stats"
)

// Record is one day of valuation data.
type Record struct {
	Date time.Time
	PER  float64
	PBR  float64
}

// CSVPath returns the conventional valuation file for a ticker.
func CSVPath(ticker string) string {
	return filepath.Join("data", ticker+"_per_pbr.csv")
}

// CSVLoader reads a Date,PER,PBR file. Extra columns are ignored, so the richer
// KRX export (EPS, BPS, DPS, dividend yield) can be used unchanged.
type CSVLoader struct{ Path string }

func (l CSVLoader) Load() ([]Record, error) {
	f, err := os.Open(l.Path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) < 2 {
		return nil, fmt.Errorf("%s: no valuation rows", l.Path)
	}

	dateCol, perCol, pbrCol := columnIndex(rows[0], "Date"), columnIndex(rows[0], "PER"), columnIndex(rows[0], "PBR")
	if dateCol < 0 || perCol < 0 || pbrCol < 0 {
		return nil, fmt.Errorf("%s: need Date, PER and PBR columns", l.Path)
	}

	var out []Record
	for _, row := range rows[1:] {
		if len(row) <= perCol || len(row) <= pbrCol {
			continue
		}
		date, err := time.Parse("2006-01-02", row[dateCol])
		if err != nil {
			continue
		}
		per, err1 := strconv.ParseFloat(row[perCol], 64)
		pbr, err2 := strconv.ParseFloat(row[pbrCol], 64)
		if err1 != nil || err2 != nil {
			continue
		}
		out = append(out, Record{Date: date, PER: per, PBR: pbr})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%s: no parseable rows", l.Path)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Date.Before(out[j].Date) })
	return out, nil
}

// YearSummary is one calendar year of valuation.
type YearSummary struct {
	Year int
	PER  stats.Summary
	PBR  stats.Summary
}

// Metric is one ratio over the window. The dates matter as much as the numbers:
// a max of 19.08 means little until you see it happened in January 2018.
type Metric struct {
	Name string
	stats.Summary
	MinDate  time.Time
	MaxDate  time.Time
	Last     float64
	LastDate time.Time
}

// Summary is the whole window.
type Summary struct {
	From, To     time.Time
	CoveredYears float64
	RequestedYrs int
	PER          Metric
	PBR          Metric
	Yearly       []YearSummary
	Records      []Record // windowed, for charting
}

// Short reports whether the data falls meaningfully short of what was asked
// for. A table headed "20년 평균" that is really six years of data is the most
// dangerous output this analysis can produce, so callers must say so.
func (s Summary) Short() bool {
	return s.CoveredYears < float64(s.RequestedYrs)-0.5
}

// Summarize restricts records to [from, to] and summarises them.
func Summarize(records []Record, from, to time.Time, requestedYears int) Summary {
	var win []Record
	for _, r := range records {
		if !r.Date.Before(from) && !r.Date.After(to) {
			win = append(win, r)
		}
	}
	if len(win) == 0 {
		return Summary{RequestedYrs: requestedYears}
	}

	per := make([]float64, len(win))
	pbr := make([]float64, len(win))
	byYear := map[int][]Record{}
	var years []int
	for i, r := range win {
		per[i], pbr[i] = r.PER, r.PBR
		if _, ok := byYear[r.Date.Year()]; !ok {
			years = append(years, r.Date.Year())
		}
		byYear[r.Date.Year()] = append(byYear[r.Date.Year()], r)
	}
	sort.Ints(years)

	yearly := make([]YearSummary, 0, len(years))
	for _, y := range years {
		rs := byYear[y]
		yp := make([]float64, len(rs))
		yb := make([]float64, len(rs))
		for i, r := range rs {
			yp[i], yb[i] = r.PER, r.PBR
		}
		yearly = append(yearly, YearSummary{Year: y, PER: stats.Summarize(yp), PBR: stats.Summarize(yb)})
	}

	last := win[len(win)-1]
	return Summary{
		From: win[0].Date, To: last.Date,
		CoveredYears: last.Date.Sub(win[0].Date).Hours() / 24 / 365.25,
		RequestedYrs: requestedYears,
		PER:          metric("PER", win, func(r Record) float64 { return r.PER }),
		PBR:          metric("PBR", win, func(r Record) float64 { return r.PBR }),
		Yearly:       yearly, Records: win,
	}
}

// metric summarises one ratio and remembers when the extremes happened.
func metric(name string, records []Record, pick func(Record) float64) Metric {
	values := make([]float64, len(records))
	minAt, maxAt := 0, 0
	for i, r := range records {
		values[i] = pick(r)
		if values[i] < values[minAt] {
			minAt = i
		}
		if values[i] > values[maxAt] {
			maxAt = i
		}
	}
	last := records[len(records)-1]
	return Metric{
		Name: name, Summary: stats.Summarize(values),
		MinDate: records[minAt].Date, MaxDate: records[maxAt].Date,
		Last: pick(last), LastDate: last.Date,
	}
}

func columnIndex(header []string, name string) int {
	for i, h := range header {
		if h == name {
			return i
		}
	}
	return -1
}

// SaveCSV writes a derived series so the next run reads the file instead of
// hitting the SEC again, and so a hand-maintained file and a derived one are
// interchangeable.
func SaveCSV(path string, records []Record) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write([]string{"Date", "PER", "PBR"}); err != nil {
		return err
	}
	for _, r := range records {
		if err := w.Write([]string{
			r.Date.Format("2006-01-02"),
			strconv.FormatFloat(r.PER, 'f', 4, 64),
			strconv.FormatFloat(r.PBR, 'f', 4, 64),
		}); err != nil {
			return err
		}
	}
	return nil
}
