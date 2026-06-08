package portfolio

type Portfolio struct {
	Shares        float64
	TotalInvested float64
	Cash          float64 // accumulated sell proceeds
	FeeRate       float64 // e.g. 0.0005 = 0.05% per transaction
}

func (p *Portfolio) Buy(amount, price float64) {
	p.Shares += amount * (1 - p.FeeRate) / price
	p.TotalInvested += amount
}

func (p *Portfolio) SellAll(price float64) {
	p.Cash += p.Shares * price * (1 - p.FeeRate)
	p.Shares = 0
}

// Reinvest converts all accumulated sell proceeds into shares without counting
// them as new invested capital. Returns false if there are no proceeds to reinvest.
func (p *Portfolio) Reinvest(price float64) bool {
	if p.Cash <= 0 {
		return false
	}
	p.Shares += p.Cash * (1 - p.FeeRate) / price
	p.Cash = 0
	return true
}

// Deposit adds KRW cash without purchasing USD (e.g. "hold KRW" baseline).
func (p *Portfolio) Deposit(amount float64) {
	p.Cash += amount
	p.TotalInvested += amount
}

func (p *Portfolio) TotalValue(price float64) float64 {
	return p.Shares*price + p.Cash
}
