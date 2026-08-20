package downloader

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type yahooResponse struct {
	Chart struct {
		Result []struct {
			Timestamp []int64 `json:"timestamp"`
			Meta      struct {
				ExchangeName         string `json:"exchangeName"`
				GMTOffset            int64  `json:"gmtoffset"`
				CurrentTradingPeriod struct {
					Regular struct {
						Start int64 `json:"start"`
						End   int64 `json:"end"`
					} `json:"regular"`
				} `json:"currentTradingPeriod"`
			} `json:"meta"`
			Indicators struct {
				Quote []struct {
					Open   []*float64 `json:"open"`
					High   []*float64 `json:"high"`
					Low    []*float64 `json:"low"`
					Close  []*float64 `json:"close"`
					Volume []*float64 `json:"volume"`
				} `json:"quote"`
				AdjClose []struct {
					AdjClose []*float64 `json:"adjclose"`
				} `json:"adjclose"`
			} `json:"indicators"`
		} `json:"result"`
	} `json:"chart"`
}

// EnsureUpToDate downloads only when the local CSV cannot already answer for
// today, so repeated runs on the same day read the file instead of the network.
func EnsureUpToDate(path string, dl func(string) error) error {
	if fresh, why := isFresh(path, time.Now()); fresh {
		fmt.Fprintf(os.Stderr, "[cache] %s: %s\n", path, why)
		return nil
	}
	return dl(path)
}

// isFresh reports whether the file already covers today, and why.
//
// Two conditions, and the second is the one that carries the weight. A CSV whose
// last row is dated today is obviously current — but that only happens after the
// exchange closes, because an unfinished session is deliberately not written.
// Before the close, and all weekend, the newest row that can exist is still
// yesterday's, so a date check alone would re-download on every run forever.
// Whether the file was fetched today answers that case correctly.
func isFresh(path string, now time.Time) (bool, string) {
	info, err := os.Stat(path)
	if err != nil {
		return false, ""
	}
	if sameDay(info.ModTime(), now) {
		return true, "오늘 받은 파일이라 다시 받지 않는다"
	}
	last, err := lastDateInCSV(path)
	if err != nil {
		return false, ""
	}
	if !last.Before(now.UTC().Truncate(24 * time.Hour)) {
		return true, "마지막 행이 오늘 날짜라 다시 받지 않는다"
	}
	return false, ""
}

func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

func lastDateInCSV(path string) (time.Time, error) {
	f, err := os.Open(path)
	if err != nil {
		return time.Time{}, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return time.Time{}, err
	}

	bufSize := int64(256)
	if bufSize > info.Size() {
		bufSize = info.Size()
	}
	buf := make([]byte, bufSize)
	if _, err := f.ReadAt(buf, info.Size()-bufSize); err != nil {
		return time.Time{}, err
	}

	s := strings.TrimRight(string(buf), "\r\n")
	if idx := strings.LastIndexByte(s, '\n'); idx >= 0 {
		s = s[idx+1:]
	}
	dateStr, _, _ := strings.Cut(s, ",")
	return time.Parse("2006-01-02", strings.TrimSpace(dateStr))
}

func TQQQ(path string) error { return Stock("TQQQ", path) }
func QLD(path string) error  { return Stock("QLD", path) }
func QQQ(path string) error  { return Stock("QQQ", path) }
func TSLA(path string) error { return Stock("TSLA", path) }
func PLTR(path string) error { return Stock("PLTR", path) }
func IREN(path string) error { return Stock("IREN", path) }
func RKLB(path string) error { return Stock("RKLB", path) }

// KIH is Korea Investment Holdings (한국금융지주), KOSPI code 071050.
// Yahoo Finance appends ".KS" for KOSPI listings (".KQ" for KOSDAQ).
// Prices are quoted in KRW, not USD.
func KIH(path string) error { return Stock("071050.KS", path) }

func Stock(symbol, path string) error {
	period1 := time.Date(1999, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	period2 := time.Now().Unix()

	url := fmt.Sprintf(
		"https://query1.finance.yahoo.com/v8/finance/chart/%s?interval=1d&period1=%d&period2=%d",
		symbol, period1, period2,
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var data yahooResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return err
	}
	if len(data.Chart.Result) == 0 {
		return fmt.Errorf("no data returned from Yahoo Finance")
	}

	result := data.Chart.Result[0]
	quote := result.Indicators.Quote[0]
	adjClose := result.Indicators.AdjClose[0].AdjClose

	loc := time.FixedZone(result.Meta.ExchangeName, int(result.Meta.GMTOffset))
	rows := make([][]string, 0, len(result.Timestamp))
	for i, ts := range result.Timestamp {
		if quote.Open[i] == nil || adjClose[i] == nil {
			continue
		}
		rows = append(rows, []string{
			time.Unix(ts, 0).In(loc).Format("2006-01-02"),
			ftos(*quote.Open[i]),
			ftos(*quote.High[i]),
			ftos(*quote.Low[i]),
			ftos(*quote.Close[i]),
			ftos(*adjClose[i]),
			strconv.FormatFloat(*quote.Volume[i], 'f', 0, 64),
		})
	}

	reg := result.Meta.CurrentTradingPeriod.Regular
	if day, partial := partialBarDate(reg.Start, reg.End, loc, time.Now()); partial && len(rows) > 0 &&
		rows[len(rows)-1][0] == day {
		fmt.Fprintf(os.Stderr, "[%s] dropping %s: the regular session is still open, so that bar is an intraday snapshot\n",
			symbol, day)
		rows = rows[:len(rows)-1]
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	w.Write([]string{"Date", "Open", "High", "Low", "Close", "AdjClose", "Volume"})
	for _, row := range rows {
		w.Write(row)
	}
	return nil
}

// partialBarDate reports the exchange-local date whose bar is still being
// written, if any.
//
// The obvious test — regularMarketTime < regular.end — is wrong. While the
// market is closed Yahoo points currentTradingPeriod at the *next* session, so
// that comparison condemns yesterday's already-final bar. A bar is only partial
// when the session it belongs to is running right now.
func partialBarDate(start, end int64, loc *time.Location, now time.Time) (string, bool) {
	if start == 0 || end == 0 {
		return "", false
	}
	if ts := now.Unix(); ts < start || ts >= end {
		return "", false
	}
	return time.Unix(start, 0).In(loc).Format("2006-01-02"), true
}

func ftos(v float64) string {
	return strconv.FormatFloat(v, 'f', 4, 64)
}

type fearGreedResponse struct {
	Historical struct {
		Data []struct {
			X float64 `json:"x"` // epoch milliseconds
			Y float64 `json:"y"` // index value
		} `json:"data"`
	} `json:"fear_and_greed_historical"`
}

// FearGreed refreshes the CNN Fear & Greed CSV in place. CNN only serves a
// rolling window of history, so existing rows are kept and only the overlap is
// rewritten: CNN restates the most recent days as its inputs settle, and the
// final point of the series is the current intraday reading, which replaces the
// day's earlier value. Rows are written back with CRLF endings to match the
// existing file.
func FearGreed(path string) error {
	dates, values, err := readFearGreedCSV(path)
	if err != nil {
		return err
	}

	// Re-fetch a margin before the last stored date so restated values are picked up.
	anchor := time.Date(2011, 1, 1, 0, 0, 0, 0, time.UTC)
	if len(dates) > 0 {
		if last, err := time.Parse("2006-01-02", dates[len(dates)-1]); err == nil {
			anchor = last.AddDate(0, 0, -30)
		}
	}

	fresh, err := fetchFearGreed(anchor)
	if err != nil {
		return err
	}

	for _, date := range fresh.dates {
		if _, ok := values[date]; !ok {
			dates = append(dates, date)
		}
		values[date] = fresh.values[date]
	}
	sort.Strings(dates)

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := fmt.Fprint(f, "Date,Fear Greed Index\r\n"); err != nil {
		return err
	}
	for _, date := range dates {
		if _, err := fmt.Fprintf(f, "%s,%s\r\n", date, strconv.FormatFloat(values[date], 'f', -1, 64)); err != nil {
			return err
		}
	}
	return nil
}

// readFearGreedCSV returns the stored dates in file order plus a date→value
// lookup. A missing file yields empty results so the CSV can be built fresh.
func readFearGreedCSV(path string) ([]string, map[string]float64, error) {
	values := make(map[string]float64)

	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, values, nil
	}
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	rows, err := r.ReadAll()
	if err != nil {
		return nil, nil, err
	}

	var dates []string
	for i, row := range rows {
		if i == 0 || len(row) < 2 {
			continue
		}
		date := strings.TrimSpace(row[0])
		v, err := strconv.ParseFloat(strings.TrimSpace(row[1]), 64)
		if err != nil {
			return nil, nil, fmt.Errorf("row %d fear greed value: %w", i+1, err)
		}
		if _, seen := values[date]; !seen {
			dates = append(dates, date)
		}
		values[date] = v
	}
	return dates, values, nil
}

type fearGreedSeries struct {
	dates  []string
	values map[string]float64
}

func fetchFearGreed(from time.Time) (fearGreedSeries, error) {
	url := "https://production.dataviz.cnn.io/index/fearandgreed/graphdata/" + from.Format("2006-01-02")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fearGreedSeries{}, err
	}
	// CNN's dataviz host rejects requests that do not look like the web client.
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Referer", "https://edition.cnn.com/")
	req.Header.Set("Origin", "https://edition.cnn.com")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fearGreedSeries{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fearGreedSeries{}, fmt.Errorf("cnn fear & greed: HTTP %d", resp.StatusCode)
	}

	var data fearGreedResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return fearGreedSeries{}, fmt.Errorf("decode cnn response: %w", err)
	}
	if len(data.Historical.Data) == 0 {
		return fearGreedSeries{}, fmt.Errorf("no data returned from CNN fear & greed")
	}

	series := fearGreedSeries{values: make(map[string]float64)}
	for _, p := range data.Historical.Data {
		date := time.UnixMilli(int64(p.X)).UTC().Format("2006-01-02")
		if _, seen := series.values[date]; !seen {
			series.dates = append(series.dates, date)
		}
		series.values[date] = p.Y
	}
	return series, nil
}

// USDKRWCombined fetches USD/KRW from Yahoo Finance and USDT/KRW from Upbit,
// then writes a single CSV with columns Date, usd_krw, usdt_krw.
// Every calendar day from start to end is included; missing usd_krw values
// (weekends/holidays) are forward-filled from the previous trading day.
// Dates with no Upbit data have an empty usdt_krw value.
func USDKRWCombined(path string, start, end time.Time) error {
	// Fetch one week before start so forward-fill has an initial value even when
	// the start date itself is a weekend or holiday.
	usdKRW, err := fetchYahooUSDKRW(start.AddDate(0, 0, -7), end)
	if err != nil {
		return fmt.Errorf("yahoo USD/KRW: %w", err)
	}

	usdtKRW, err := fetchUpbitUSDTKRW(start, end)
	if err != nil {
		return fmt.Errorf("upbit USDT/KRW: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	w.Write([]string{"Date", "usd_krw", "usdt_krw"})

	// Seed lastUSD from the lookback window before the requested start date.
	var lastUSD float64
	for d := start.AddDate(0, 0, -7); d.Before(start); d = d.AddDate(0, 0, 1) {
		if v, ok := usdKRW[d.Format("2006-01-02")]; ok {
			lastUSD = v
		}
	}

	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		date := d.Format("2006-01-02")
		if v, ok := usdKRW[date]; ok {
			lastUSD = v
		}
		usdt := ""
		if v, ok := usdtKRW[date]; ok {
			usdt = ftos(v)
		}
		if lastUSD != 0 {
			w.Write([]string{date, ftos(lastUSD), usdt})
		}
	}
	return nil
}

func fetchYahooUSDKRW(start, end time.Time) (prices map[string]float64, err error) {
	url := fmt.Sprintf(
		"https://query1.finance.yahoo.com/v8/finance/chart/KRW%%3DX?interval=1d&period1=%d&period2=%d",
		start.Unix(), end.Unix(),
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data yahooResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	if len(data.Chart.Result) == 0 {
		return nil, fmt.Errorf("no data returned from Yahoo Finance for KRW=X")
	}

	result := data.Chart.Result[0]
	adjCloseSlice := result.Indicators.AdjClose[0].AdjClose
	prices = make(map[string]float64)

	for i, ts := range result.Timestamp {
		if adjCloseSlice[i] == nil {
			continue
		}
		date := time.Unix(ts, 0).UTC()
		if date.Before(start) || date.After(end) {
			continue
		}
		prices[date.Format("2006-01-02")] = *adjCloseSlice[i]
	}
	return prices, nil
}

type upbitCandle struct {
	DateTimeUTC string  `json:"candle_date_time_utc"`
	TradePrice  float64 `json:"trade_price"`
}

func fetchUpbitUSDTKRW(start, end time.Time) (map[string]float64, error) {
	prices := make(map[string]float64)
	cursor := end.AddDate(0, 0, 1)

	for {
		candles, err := fetchUpbitCandles(cursor, 200)
		if err != nil {
			return nil, err
		}
		if len(candles) == 0 {
			break
		}

		done := false
		for _, c := range candles {
			t, err := time.Parse("2006-01-02T15:04:05", c.DateTimeUTC)
			if err != nil {
				return nil, fmt.Errorf("parse upbit date %q: %w", c.DateTimeUTC, err)
			}
			if t.Before(start) {
				done = true
				break
			}
			prices[t.Format("2006-01-02")] = c.TradePrice
		}

		if done || len(candles) < 200 {
			break
		}

		oldest, _ := time.Parse("2006-01-02T15:04:05", candles[len(candles)-1].DateTimeUTC)
		cursor = oldest
	}
	return prices, nil
}

func fetchUpbitCandles(to time.Time, count int) ([]upbitCandle, error) {
	url := fmt.Sprintf(
		"https://api.upbit.com/v1/candles/days?market=KRW-USDT&count=%d&to=%s",
		count, to.UTC().Format("2006-01-02T15:04:05Z"),
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var candles []upbitCandle
	if err := json.NewDecoder(resp.Body).Decode(&candles); err != nil {
		return nil, fmt.Errorf("decode upbit response: %w", err)
	}
	return candles, nil
}
