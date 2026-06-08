package strategy

import (
	"math"
	"stock-btc-backtest/portfolio"
	"time"
)

// KimchiStrategy deposits InitialAmount KRW on day 1, then cycles that capital:
//   - Buys (Reinvest) all available KRW cash at USDT/KRW price when
//     indicator <= BuyThreshold, at most once per month.
//   - Sells all USDT when indicator >= SellThreshold.
//
// "current cash" = the KRW received from the previous sell (0 if never sold).
// No new money is injected after the initial deposit.
// The indicator is PremiumPct (%) or PremiumWon (₩) depending on the backtester.
type KimchiStrategy struct {
	Label         string
	InitialAmount float64
	BuyThreshold  float64
	SellThreshold float64 // 0 = never sell

	started    bool
	lastBuyKey int
}

func (s *KimchiStrategy) Name() string { return s.Label }

func (s *KimchiStrategy) OnDay(date time.Time, usdtKRW float64, indicator float64, p *portfolio.Portfolio) {
	if !s.started {
		p.Deposit(s.InitialAmount)
		s.started = true
		return
	}

	if math.IsNaN(indicator) {
		return
	}

	mk := monthKey(date.Year(), date.Month())

	if s.SellThreshold != 0 && indicator >= s.SellThreshold && p.Shares > 0 {
		p.SellAll(usdtKRW) // sell USDT at USDT/KRW price
		return
	}

	if indicator <= s.BuyThreshold && mk > s.lastBuyKey {
		if p.Reinvest(usdtKRW) { // buy USDT at USDT/KRW price, only if cash available
			s.lastBuyKey = mk
		}
	}
}

// USDKRWStrategy deposits InitialAmount KRW on day 1, then cycles that capital:
//   - Trigger is USD/KRW rate (indicator param, from rateMap).
//   - Actual buy/sell transactions execute at USDT/KRW price (price param).
//   - Buys (Reinvest) when USD/KRW <= BuyRate, at most once per month.
//   - Sells all USDT when USD/KRW >= SellRate.
type USDKRWStrategy struct {
	Label         string
	InitialAmount float64
	BuyRate       float64
	SellRate      float64

	started    bool
	lastBuyKey int
}

func (s *USDKRWStrategy) Name() string { return s.Label }

func (s *USDKRWStrategy) OnDay(date time.Time, usdtKRW float64, usdKRW float64, p *portfolio.Portfolio) {
	if !s.started {
		p.Deposit(s.InitialAmount)
		s.started = true
		return
	}

	mk := monthKey(date.Year(), date.Month())

	if usdKRW >= s.SellRate && p.Shares > 0 {
		p.SellAll(usdtKRW) // trigger: USD/KRW rate; transaction: USDT/KRW price
		return
	}

	if usdKRW <= s.BuyRate && mk > s.lastBuyKey {
		if p.Reinvest(usdtKRW) { // trigger: USD/KRW rate; transaction: USDT/KRW price
			s.lastBuyKey = mk
		}
	}
}
