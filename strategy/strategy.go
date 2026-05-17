package strategy

import (
	"stock-btc-backtest/portfolio"
	"time"
)

type Strategy interface {
	Name() string
	OnDay(date time.Time, price float64, fearGreed float64, p *portfolio.Portfolio)
}

func monthKey(year int, month time.Month) int {
	return year*12 + int(month)
}

// LumpSumStrategy: invest the full amount on the very first trading day.
// Represents Strategy 1.
type LumpSumStrategy struct {
	Label  string
	Amount float64
	done   bool
}

func (s *LumpSumStrategy) Name() string { return s.Label }

func (s *LumpSumStrategy) OnDay(_ time.Time, price float64, _ float64, p *portfolio.Portfolio) {
	if !s.done {
		p.Buy(s.Amount, price)
		s.done = true
	}
}

// AnnualDCAStrategy: lump-sum initial investment + fixed amount every January.
// Represents Strategy 2.
type AnnualDCAStrategy struct {
	InitialAmount float64
	AnnualAmount  float64
	lastBuyYear   int
	started       bool
}

func (s *AnnualDCAStrategy) Name() string {
	return "Strategy 2: $12,000 every Jan 1st"
}

func (s *AnnualDCAStrategy) OnDay(date time.Time, price float64, _ float64, p *portfolio.Portfolio) {
	if !s.started {
		p.Buy(s.InitialAmount, price)
		s.started = true
		s.lastBuyYear = date.Year() // skip annual buy in the same year as start
		return
	}
	if date.Month() == time.January && date.Year() > s.lastBuyYear {
		p.Buy(s.AnnualAmount, price)
		s.lastBuyYear = date.Year()
	}
}

// MonthlyDCAStrategy: lump-sum initial + fixed amount on or after a given day each month.
// Represents Strategy 2 (day=25) and Strategy 3 (day=1).
type MonthlyDCAStrategy struct {
	Label         string
	InitialAmount float64
	MonthlyAmount float64
	DayOfMonth    int
	started       bool
	lastBuyKey    int
}

func (s *MonthlyDCAStrategy) Name() string { return s.Label }

func (s *MonthlyDCAStrategy) OnDay(date time.Time, price float64, _ float64, p *portfolio.Portfolio) {
	if !s.started {
		p.Buy(s.InitialAmount, price)
		s.started = true
		return
	}
	if date.Day() >= s.DayOfMonth {
		key := monthKey(date.Year(), date.Month())
		if key > s.lastBuyKey {
			p.Buy(s.MonthlyAmount, price)
			s.lastBuyKey = key
		}
	}
}

// FearGreedTieredStrategy: accumulates $1,000/month and deploys a percentage of
// all saved cash based on how fearful the market is (once per month).
//
//	F&G <= 10 → buy 100% of cash
//	F&G <= 15 → buy  80% of cash
//	F&G <= 24 → buy  50% of cash
type FearGreedTieredStrategy struct {
	Label         string
	MonthlyAmount float64

	started      bool
	pendingCash  float64
	lastAccumKey int
	lastBuyKey   int
}

func (s *FearGreedTieredStrategy) Name() string { return s.Label }

func (s *FearGreedTieredStrategy) OnDay(date time.Time, price float64, fearGreed float64, p *portfolio.Portfolio) {
	if !s.started {
		s.started = true
		mk := monthKey(date.Year(), date.Month())
		s.lastBuyKey = mk
		s.lastAccumKey = mk - 1
		return
	}

	mk := monthKey(date.Year(), date.Month())

	if mk > s.lastAccumKey {
		s.pendingCash += s.MonthlyAmount
		s.lastAccumKey = mk
	}

	if fearGreed < 0 || s.pendingCash <= 0 || mk <= s.lastBuyKey {
		return
	}

	var pct float64
	switch {
	case fearGreed <= 10:
		pct = 1.0
	case fearGreed <= 15:
		pct = 0.8
	case fearGreed <= 24:
		pct = 0.5
	default:
		return
	}

	amount := s.pendingCash * pct
	p.Buy(amount, price)
	s.pendingCash -= amount
	s.lastBuyKey = mk
}

// FearGreedAccumStrategy: invests an initial lump sum on day 1, then accumulates
// a fixed monthly amount as cash. On extreme fear, deploys ALL available cash
// (savings + any sell proceeds) at once. Optionally sells everything on extreme
// greed when SellThreshold > 0.
type FearGreedAccumStrategy struct {
	Label         string
	InitialAmount float64 // invested on day 1
	MonthlyAmount float64 // added to cash each non-buy month
	BuyThreshold  float64 // F&G <= this triggers buy-all
	SellThreshold float64 // F&G >= this triggers sell-all (0 = never sell)

	started      bool
	pendingCash  float64 // monthly savings not yet deployed
	lastAccumKey int
	lastBuyKey   int
}

func (s *FearGreedAccumStrategy) Name() string { return s.Label }

func (s *FearGreedAccumStrategy) OnDay(date time.Time, price float64, fearGreed float64, p *portfolio.Portfolio) {
	if !s.started {
		p.Buy(s.InitialAmount, price)
		s.started = true
		mk := monthKey(date.Year(), date.Month())
		s.lastBuyKey = mk
		s.lastAccumKey = mk - 1 // allow accumulation to start in the first month
		return
	}

	mk := monthKey(date.Year(), date.Month())

	// Accumulate on the first trading day of each new month (always, even in sell months)
	if mk > s.lastAccumKey {
		s.pendingCash += s.MonthlyAmount
		s.lastAccumKey = mk
	}

	// Sell all on greed
	if s.SellThreshold > 0 && fearGreed >= s.SellThreshold && p.Shares > 0 {
		p.SellAll(price)
		return
	}

	// Deploy all cash on extreme fear (once per month)
	if fearGreed >= 0 && fearGreed <= s.BuyThreshold && mk > s.lastBuyKey {
		if s.pendingCash > 0 {
			p.Buy(s.pendingCash, price) // new capital → tracked in TotalInvested
			s.pendingCash = 0
		}
		p.Reinvest(price) // sell proceeds → not tracked in TotalInvested
		s.lastBuyKey = mk
	}
}
