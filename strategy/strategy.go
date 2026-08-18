package strategy

import (
	"stock-btc-backtest/portfolio"
	"time"
)

type Strategy interface {
	Name() string
	OnDay(date time.Time, price float64, indicator float64, p *portfolio.Portfolio)
}

func monthKey(year int, month time.Month) int {
	return year*12 + int(month)
}

// LumpSumStrategy: invest the full amount on the very first trading day.
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

// AnnualDCAStrategy: fixed amount every January 1st.
type AnnualDCAStrategy struct {
	Label         string
	InitialAmount float64
	AnnualAmount  float64
	lastBuyYear   int
	started       bool
}

func (s *AnnualDCAStrategy) Name() string {
	if s.Label != "" {
		return s.Label
	}
	return "Strategy 2: $12,000 every Jan 1st"
}

func (s *AnnualDCAStrategy) OnDay(date time.Time, price float64, _ float64, p *portfolio.Portfolio) {
	if !s.started {
		p.Buy(s.InitialAmount, price)
		s.started = true
		s.lastBuyYear = date.Year()
		return
	}
	if date.Month() == time.January && date.Year() > s.lastBuyYear {
		p.Buy(s.AnnualAmount, price)
		s.lastBuyYear = date.Year()
	}
}

// MonthlyDCAStrategy: fixed amount on or after a given day each month.
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
