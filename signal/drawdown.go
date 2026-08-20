package signal

import (
	"fmt"
	"math"
)

// DrawdownAtLeast is met when the price has fallen at least Pct below the
// running peak. Pct is a positive fraction: 0.30 means 30% off the high.
type DrawdownAtLeast struct{ Pct float64 }

func (c DrawdownAtLeast) Met(f Frame, i int) bool {
	dd := f.DrawdownPct(i)
	return !math.IsNaN(dd) && dd <= -c.Pct*100
}

func (c DrawdownAtLeast) Describe() string {
	return fmt.Sprintf("연중 러닝 고점 대비 -%.1f%% 이하", c.Pct*100)
}
