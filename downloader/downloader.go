package downloader

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

type yahooResponse struct {
	Chart struct {
		Result []struct {
			Timestamp  []int64 `json:"timestamp"`
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

func TQQQ(path string) error { return Stock("TQQQ", path) }
func QLD(path string) error  { return Stock("QLD", path) }
func QQQ(path string) error  { return Stock("QQQ", path) }

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

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	w.Write([]string{"Date", "Open", "High", "Low", "Close", "AdjClose", "Volume"})

	for i, ts := range result.Timestamp {
		if quote.Open[i] == nil || adjClose[i] == nil {
			continue
		}
		w.Write([]string{
			time.Unix(ts, 0).UTC().Format("2006-01-02"),
			ftos(*quote.Open[i]),
			ftos(*quote.High[i]),
			ftos(*quote.Low[i]),
			ftos(*quote.Close[i]),
			ftos(*adjClose[i]),
			strconv.FormatFloat(*quote.Volume[i], 'f', 0, 64),
		})
	}
	return nil
}

func ftos(v float64) string {
	return strconv.FormatFloat(v, 'f', 4, 64)
}
