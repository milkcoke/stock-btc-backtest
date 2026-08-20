package valuation

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// EDGAR reads reported fundamentals from the SEC's XBRL API, which is free,
// needs no key, and reaches back to roughly 2009 for most filers.
//
// It only covers SEC filers, so non-US listings (a .KS ticker, say) will not
// resolve. That is a clean miss rather than a wrong answer: callers fall back to
// a hand-maintained CSV.
type EDGAR struct {
	// UserAgent identifies the caller. data.sec.gov rejects requests without one.
	UserAgent string
}

const defaultEDGARUserAgent = "stock-btc-backtest research script"

func (e EDGAR) userAgent() string {
	if e.UserAgent == "" {
		return defaultEDGARUserAgent
	}
	return e.UserAgent
}

func (e EDGAR) Fetch(ticker string) (Fundamentals, error) {
	cik, err := e.resolveCIK(ticker)
	if err != nil {
		return Fundamentals{}, err
	}

	eps, err := e.concept(cik, "us-gaap", "EarningsPerShareDiluted", "USD/shares")
	if err != nil {
		// Some filers only tag the basic figure.
		eps, err = e.concept(cik, "us-gaap", "EarningsPerShareBasic", "USD/shares")
		if err != nil {
			return Fundamentals{}, fmt.Errorf("%s: no diluted or basic EPS reported: %w", ticker, err)
		}
	}

	equity, err := e.concept(cik, "us-gaap", "StockholdersEquity", "USD")
	if err != nil {
		return Fundamentals{}, fmt.Errorf("%s: no stockholders' equity reported: %w", ticker, err)
	}
	shares, err := e.concept(cik, "dei", "EntityCommonStockSharesOutstanding", "shares")
	if err != nil {
		return Fundamentals{}, fmt.Errorf("%s: no share count reported: %w", ticker, err)
	}

	quarterly, annual := splitByDuration(eps)
	return Fundamentals{
		QuarterlyEPS: quarterly, AnnualEPS: annual,
		Equity: equity, Shares: shares,
	}, nil
}

// resolveCIK maps a ticker to its SEC identifier.
//
// The obvious file for this, www.sec.gov/files/company_tickers.json, answers 403
// to scripted clients, so this goes through full-text search and then confirms
// the hit against the company's own submissions record. The confirmation matters:
// a search for a short ticker easily surfaces some other company that merely
// mentioned it.
func (e EDGAR) resolveCIK(ticker string) (string, error) {
	endpoint := "https://efts.sec.gov/LATEST/search-index?q=%22" + url.QueryEscape(ticker) + "%22&forms=10-K"
	var payload struct {
		Aggregations struct {
			EntityFilter struct {
				Buckets []struct {
					Key string `json:"key"`
				} `json:"buckets"`
			} `json:"entity_filter"`
		} `json:"aggregations"`
	}
	if err := e.getJSON(endpoint, &payload); err != nil {
		return "", err
	}

	// Buckets look like: "TJX COMPANIES INC /DE/  (TJX)  (CIK 0000109198)"
	pattern := regexp.MustCompile(`\(([^()]*)\)\s*\(CIK (\d{10})\)`)
	for _, b := range payload.Aggregations.EntityFilter.Buckets {
		m := pattern.FindStringSubmatch(b.Key)
		if m == nil {
			continue
		}
		if !containsTicker(m[1], ticker) {
			continue
		}
		if e.confirmTicker(m[2], ticker) {
			return m[2], nil
		}
	}
	return "", fmt.Errorf("%s: no SEC filer found (non-US listings have no EDGAR data)", ticker)
}

func (e EDGAR) confirmTicker(cik, ticker string) bool {
	var sub struct {
		Tickers []string `json:"tickers"`
	}
	if err := e.getJSON("https://data.sec.gov/submissions/CIK"+cik+".json", &sub); err != nil {
		return false
	}
	for _, t := range sub.Tickers {
		if strings.EqualFold(t, ticker) {
			return true
		}
	}
	return false
}

func containsTicker(list, ticker string) bool {
	for _, part := range strings.Split(list, ",") {
		if strings.EqualFold(strings.TrimSpace(part), ticker) {
			return true
		}
	}
	return false
}

type conceptResponse struct {
	Units map[string][]struct {
		Start string  `json:"start"`
		End   string  `json:"end"`
		Filed string  `json:"filed"`
		Val   float64 `json:"val"`
	} `json:"units"`
}

func (e EDGAR) concept(cik, namespace, name, unit string) ([]Report, error) {
	endpoint := fmt.Sprintf("https://data.sec.gov/api/xbrl/companyconcept/CIK%s/%s/%s.json", cik, namespace, name)
	var payload conceptResponse
	if err := e.getJSON(endpoint, &payload); err != nil {
		return nil, err
	}

	facts, ok := payload.Units[unit]
	if !ok {
		return nil, fmt.Errorf("unit %s not reported for %s", unit, name)
	}

	// Keep the earliest filing for each period. Later filings repeat old periods
	// as comparatives; taking those would date the figure years after the market
	// actually saw it.
	earliest := map[string]Report{}
	for _, f := range facts {
		end, err := time.Parse("2006-01-02", f.End)
		if err != nil {
			continue
		}
		filed, err := time.Parse("2006-01-02", f.Filed)
		if err != nil {
			continue
		}
		key := f.Start + "|" + f.End
		if prev, seen := earliest[key]; seen && !filed.Before(prev.Filed) {
			continue
		}
		r := Report{End: end, Filed: filed, Value: f.Val}
		if f.Start != "" {
			if start, err := time.Parse("2006-01-02", f.Start); err == nil {
				r.periodDays = int(end.Sub(start).Hours() / 24)
			}
		}
		earliest[key] = r
	}

	out := make([]Report, 0, len(earliest))
	for _, r := range earliest {
		out = append(out, r)
	}
	sortReports(out)
	return out, nil
}

func (e EDGAR) getJSON(endpoint string, into any) error {
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", e.userAgent())
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: HTTP %d", endpoint, resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(into)
}

// splitByDuration separates three-month periods from full years. Issuers that
// fold the fourth quarter into the annual report leave a gap in the quarterly
// series, and the annual figure fills it.
func splitByDuration(reports []Report) (quarterly, annual []Report) {
	for _, r := range reports {
		switch {
		case r.periodDays >= 80 && r.periodDays <= 100:
			quarterly = append(quarterly, r)
		case r.periodDays >= 350 && r.periodDays <= 380:
			annual = append(annual, r)
		}
	}
	return quarterly, annual
}
