package backtester

import (
	"math"
	"time"

	"stock-btc-backtest/loader"
	"stock-btc-backtest/portfolio"
	"stock-btc-backtest/strategy"
)

type DataPoint struct {
	Date  string  // "2006-01"
	Value float64 // total portfolio value in USD
}

type Result struct {
	StrategyName  string
	TotalInvested float64
	FinalValue    float64
	MDD           float64 // Maximum Drawdown in percent (negative value)
	MDDDate       string  // month when the trough occurred (YYYY-MM)
	History       []DataPoint
}

func (r Result) ReturnPct() float64 {
	if r.TotalInvested == 0 {
		return 0
	}
	return (r.FinalValue - r.TotalInvested) / r.TotalInvested * 100
}

type Backtester struct {
	prices       []loader.PriceRecord
	fearGreed    map[string]float64
	extraDailyER float64 // (thisETF_TER - priceData_TER) / 365; negative means lower fees than price data
}

func New(prices []loader.PriceRecord, fearGreed map[string]float64, annualERDelta float64) *Backtester {
	return &Backtester{prices: prices, fearGreed: fearGreed, extraDailyER: annualERDelta / 365.0}
}

func (b *Backtester) Run(s strategy.Strategy, start, end time.Time) Result {
	p := &portfolio.Portfolio{}
	var lastPrice, peak, mdd float64
	var mddDate string
	var history []DataPoint
	var lastMonthKey int

	for _, pr := range b.prices {
		if pr.Date.Before(start) || pr.Date.After(end) {
			continue
		}
		fg, ok := b.fearGreed[pr.Date.Format("2006-01-02")]
		if !ok {
			fg = math.NaN()
		}
		s.OnDay(pr.Date, pr.AdjClose, fg, p)
		if b.extraDailyER != 0 {
			p.Shares *= (1 - b.extraDailyER)
		}
		lastPrice = pr.AdjClose

		value := p.TotalValue(pr.AdjClose)
		if value > peak {
			peak = value
		}
		if peak > 0 {
			if dd := (value - peak) / peak * 100; dd < mdd {
				mdd = dd
				mddDate = pr.Date.Format("2006-01")
			}
		}

		if mk := pr.Date.Year()*12 + int(pr.Date.Month()); mk != lastMonthKey {
			history = append(history, DataPoint{
				Date:  pr.Date.Format("2006-01"),
				Value: value,
			})
			lastMonthKey = mk
		}
	}

	return Result{
		StrategyName:  s.Name(),
		TotalInvested: p.TotalInvested,
		FinalValue:    p.TotalValue(lastPrice),
		MDD:           mdd,
		MDDDate:       mddDate,
		History:       history,
	}
}
