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
	TradeCount    int     // total buy + sell events
	AvgHoldDays   float64 // avg days from position open to close; 0 if no complete cycles
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
	tradingFee   float64 // fraction deducted on each buy/sell (e.g. 0.0005 = 0.05%)
}

func New(prices []loader.PriceRecord, fearGreed map[string]float64, annualERDelta float64) *Backtester {
	return &Backtester{prices: prices, fearGreed: fearGreed, extraDailyER: annualERDelta / 365.0}
}

func (b *Backtester) WithTradingFee(fee float64) *Backtester {
	b.tradingFee = fee
	return b
}

func (b *Backtester) Run(s strategy.Strategy, start, end time.Time) Result {
	p := &portfolio.Portfolio{FeeRate: b.tradingFee}
	var lastPrice, peak, mdd float64
	var mddDate string
	var history []DataPoint
	var lastMonthKey int

	var tradeCount int
	var positionOpenDate time.Time
	var totalHoldDays float64
	var completeCycles int
	const eps = 1e-9

	for _, pr := range b.prices {
		if pr.Date.Before(start) || pr.Date.After(end) {
			continue
		}
		fg, ok := b.fearGreed[pr.Date.Format("2006-01-02")]
		if !ok {
			fg = math.NaN()
		}

		sharesBefore := p.Shares
		s.OnDay(pr.Date, pr.AdjClose, fg, p)
		sharesAfter := p.Shares // capture before extraDailyER (which is not a trade)

		if b.extraDailyER != 0 {
			p.Shares *= (1 - b.extraDailyER)
		}
		lastPrice = pr.AdjClose

		// Detect buy: shares increased
		if sharesAfter > sharesBefore+eps {
			tradeCount++
			if positionOpenDate.IsZero() {
				positionOpenDate = pr.Date
			}
		}
		// Detect sell: shares cleared to zero
		if sharesAfter < eps && sharesBefore > eps {
			tradeCount++
			if !positionOpenDate.IsZero() {
				totalHoldDays += pr.Date.Sub(positionOpenDate).Hours() / 24
				completeCycles++
				positionOpenDate = time.Time{}
			}
		}

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

	var avgHoldDays float64
	if completeCycles > 0 {
		avgHoldDays = totalHoldDays / float64(completeCycles)
	}

	return Result{
		StrategyName:  s.Name(),
		TotalInvested: p.TotalInvested,
		FinalValue:    p.TotalValue(lastPrice),
		MDD:           mdd,
		MDDDate:       mddDate,
		History:       history,
		TradeCount:    tradeCount,
		AvgHoldDays:   avgHoldDays,
	}
}
