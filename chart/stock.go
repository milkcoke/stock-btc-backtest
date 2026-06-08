package chart

import "time"

func GenerateStock(outputPath string, tickers []TickerChart, start, end time.Time) error {
	return Generate(outputPath, tickers, start, end)
}

func GenerateStockComparison(outputPath, strategyName string, lines []ComparisonLine, start, end time.Time) error {
	return GenerateComparison(outputPath, strategyName, lines, start, end)
}
