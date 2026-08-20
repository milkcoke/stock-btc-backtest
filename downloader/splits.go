package downloader

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SplitEvent is one stock split. Ratio is the multiplier applied to the share
// count: a 3-for-1 split has Ratio 3.
type SplitEvent struct {
	Date  time.Time
	Ratio float64
}

type splitResponse struct {
	Chart struct {
		Result []struct {
			Meta struct {
				GMTOffset int64 `json:"gmtoffset"`
			} `json:"meta"`
			Events struct {
				Splits map[string]struct {
					Date        int64   `json:"date"`
					Numerator   float64 `json:"numerator"`
					Denominator float64 `json:"denominator"`
				} `json:"splits"`
			} `json:"events"`
		} `json:"result"`
	} `json:"chart"`
}

// SplitEvents returns every split Yahoo knows about, oldest first.
//
// Price CSVs are back-adjusted for splits but per-share figures filed with a
// regulator are not, so anything that divides a price by a reported per-share
// number needs this list to put both on the same basis.
func SplitEvents(symbol string) ([]SplitEvent, error) {
	url := fmt.Sprintf(
		"https://query1.finance.yahoo.com/v8/finance/chart/%s?interval=1d&range=max&events=split", symbol)
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

	var data splitResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	if len(data.Chart.Result) == 0 {
		return nil, fmt.Errorf("%s: no split data returned", symbol)
	}

	result := data.Chart.Result[0]
	loc := time.FixedZone("exchange", int(result.Meta.GMTOffset))
	out := make([]SplitEvent, 0, len(result.Events.Splits))
	for _, s := range result.Events.Splits {
		if s.Denominator == 0 {
			continue
		}
		out = append(out, SplitEvent{
			Date:  time.Unix(s.Date, 0).In(loc).Truncate(24 * time.Hour),
			Ratio: s.Numerator / s.Denominator,
		})
	}
	sortSplits(out)
	return out, nil
}

func sortSplits(s []SplitEvent) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j].Date.Before(s[j-1].Date); j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}
