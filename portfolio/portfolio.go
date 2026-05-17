package portfolio

type Portfolio struct {
	Shares        float64
	TotalInvested float64
	Cash          float64 // accumulated sell proceeds
}

func (p *Portfolio) Buy(amount, price float64) {
	p.Shares += amount / price
	p.TotalInvested += amount
}

func (p *Portfolio) SellAll(price float64) {
	p.Cash += p.Shares * price
	p.Shares = 0
}

// Reinvest converts all accumulated sell proceeds into shares without counting
// them as new invested capital. Returns false if there are no proceeds to reinvest.
func (p *Portfolio) Reinvest(price float64) bool {
	if p.Cash <= 0 {
		return false
	}
	p.Shares += p.Cash / price
	p.Cash = 0
	return true
}

func (p *Portfolio) TotalValue(price float64) float64 {
	return p.Shares*price + p.Cash
}
