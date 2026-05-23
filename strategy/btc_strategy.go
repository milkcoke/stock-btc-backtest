package strategy

import (
	"math"
	"stock-btc-backtest/portfolio"
	"time"
)

// MVRVAccumStrategy: accumulates $1,000/month and deploys all cash when BTC
// MVRV Z-Score drops to or below BuyThreshold (including negative values).
// Optionally sells all when MVRV rises to or above SellThreshold (0 = never sell).
type MVRVAccumStrategy struct {
	Label         string
	MonthlyAmount float64
	BuyThreshold  float64
	SellThreshold float64 // 0 = never sell

	started      bool
	pendingCash  float64
	lastAccumKey int
	lastBuyKey   int
}

func (s *MVRVAccumStrategy) Name() string { return s.Label }

func (s *MVRVAccumStrategy) OnDay(date time.Time, price float64, mvrvZ float64, p *portfolio.Portfolio) {
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

	if math.IsNaN(mvrvZ) {
		return
	}

	if s.SellThreshold != 0 && mvrvZ >= s.SellThreshold && p.Shares > 0 {
		p.SellAll(price)
		return
	}

	if mvrvZ <= s.BuyThreshold && mk > s.lastBuyKey {
		if s.pendingCash > 0 {
			p.Buy(s.pendingCash, price)
			s.pendingCash = 0
		}
		p.Reinvest(price)
		s.lastBuyKey = mk
	}
}
