package signal

import (
	"fmt"
	"math"
)

// RSIBelow is met when RSI is at or under Max.
type RSIBelow struct {
	Max    float64
	Period int // reporting only
}

func (c RSIBelow) Met(f Frame, i int) bool {
	return !math.IsNaN(f.RSI[i]) && f.RSI[i] <= c.Max
}

func (c RSIBelow) Describe() string {
	return fmt.Sprintf("RSI(%d) ≤ %.0f", c.Period, c.Max)
}
