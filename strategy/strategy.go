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

// FearGreedBuyStrategy: buy a fixed amount when F&G index is at or below threshold,
// at most once per calendar month. Represents Strategy 4.
type FearGreedBuyStrategy struct {
	Label        string
	BuyAmount    float64
	BuyThreshold float64
	lastBuyKey   int
}

func (s *FearGreedBuyStrategy) Name() string { return s.Label }

func (s *FearGreedBuyStrategy) OnDay(date time.Time, price float64, fearGreed float64, p *portfolio.Portfolio) {
	if fearGreed < 0 || fearGreed > s.BuyThreshold {
		return
	}
	key := monthKey(date.Year(), date.Month())
	if key > s.lastBuyKey {
		p.Buy(s.BuyAmount, price)
		s.lastBuyKey = key
	}
}

// FearGreedBuySellStrategy: buy on extreme fear (once/month), sell all on extreme greed.
// Represents Strategy 5.
type FearGreedBuySellStrategy struct {
	Label         string
	BuyAmount     float64
	BuyThreshold  float64
	SellThreshold float64
	lastBuyKey    int
}

func (s *FearGreedBuySellStrategy) Name() string { return s.Label }

func (s *FearGreedBuySellStrategy) OnDay(date time.Time, price float64, fearGreed float64, p *portfolio.Portfolio) {
	if fearGreed < 0 {
		return
	}
	if fearGreed >= s.SellThreshold && p.Shares > 0 {
		p.SellAll(price)
		return
	}
	if fearGreed <= s.BuyThreshold {
		key := monthKey(date.Year(), date.Month())
		if key > s.lastBuyKey {
			if !p.Reinvest(price) {
				p.Buy(s.BuyAmount, price) // first buy: inject new capital
			}
			s.lastBuyKey = key
		}
	}
}
