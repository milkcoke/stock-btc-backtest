package strategy

import (
	"math"
	"stock-btc-backtest/portfolio"
	"time"
)

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

	if math.IsNaN(fearGreed) || s.pendingCash <= 0 || mk <= s.lastBuyKey {
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

// FearGreedAccumStrategy: accumulates a fixed monthly amount as cash.
// On extreme fear (F&G <= BuyThreshold), deploys ALL available cash at once.
// Optionally sells everything on extreme greed when SellThreshold > 0.
type FearGreedAccumStrategy struct {
	Label         string
	InitialAmount float64
	MonthlyAmount float64
	BuyThreshold  float64 // F&G <= this triggers buy-all
	SellThreshold float64 // F&G >= this triggers sell-all (0 = never sell)

	started      bool
	pendingCash  float64
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
		s.lastAccumKey = mk - 1
		return
	}

	mk := monthKey(date.Year(), date.Month())

	if mk > s.lastAccumKey {
		s.pendingCash += s.MonthlyAmount
		s.lastAccumKey = mk
	}

	if s.SellThreshold > 0 && fearGreed >= s.SellThreshold && p.Shares > 0 {
		p.SellAll(price)
		return
	}

	if fearGreed >= 0 && fearGreed <= s.BuyThreshold && mk > s.lastBuyKey {
		if s.pendingCash > 0 {
			p.Buy(s.pendingCash, price)
			s.pendingCash = 0
		}
		p.Reinvest(price)
		s.lastBuyKey = mk
	}
}
