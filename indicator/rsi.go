package indicator

import "math"

// WilderRSI returns the Relative Strength Index using Wilder's smoothing, the
// definition charting platforms use. The first period positions are NaN.
//
// The seed is the simple average of the first period gains and losses; every
// later step blends the new value in at 1/period weight. A zero average loss
// yields 100 rather than dividing by zero.
func WilderRSI(values []float64, period int) []float64 {
	out := make([]float64, len(values))
	for i := range out {
		out[i] = math.NaN()
	}
	if period <= 0 || len(values) <= period {
		return out
	}

	var avgGain, avgLoss float64
	for i := 1; i <= period; i++ {
		gain, loss := change(values[i-1], values[i])
		avgGain += gain
		avgLoss += loss
	}
	avgGain /= float64(period)
	avgLoss /= float64(period)
	out[period] = rsiFrom(avgGain, avgLoss)

	for i := period + 1; i < len(values); i++ {
		gain, loss := change(values[i-1], values[i])
		avgGain = (avgGain*float64(period-1) + gain) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + loss) / float64(period)
		out[i] = rsiFrom(avgGain, avgLoss)
	}
	return out
}

func change(prev, cur float64) (gain, loss float64) {
	if d := cur - prev; d > 0 {
		return d, 0
	} else {
		return 0, -d
	}
}

func rsiFrom(avgGain, avgLoss float64) float64 {
	if avgLoss == 0 {
		return 100
	}
	return 100 - 100/(1+avgGain/avgLoss)
}
