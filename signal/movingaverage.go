package signal

import (
	"fmt"
	"math"
)

// MADiscount is met when the price trades at least Pct below its moving
// average. Pct is a positive fraction: 0.10 means a 10% discount.
type MADiscount struct {
	Pct    float64
	Window int // reporting only; the frame already carries the right SMA
}

func (c MADiscount) Met(f Frame, i int) bool {
	d := f.MADiscountPct(i)
	return !math.IsNaN(d) && d <= -c.Pct*100
}

func (c MADiscount) Describe() string {
	return fmt.Sprintf("MA%d 대비 -%.0f%% 이하", c.Window, c.Pct*100)
}
