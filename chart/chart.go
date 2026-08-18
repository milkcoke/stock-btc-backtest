package chart

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"strings"
	"time"

	"stock-btc-backtest/backtester"
)

var palette = []string{
	"#e74c3c", "#e67e22", "#f1c40f", "#2ecc71",
	"#3498db", "#9b59b6", "#1abc9c",
}

type TickerChart struct {
	Symbol        string
	Results       []backtester.Result
	Currency      string // symbol prefix, e.g. "$" or "₩" (default "$")
	CurrencyLabel string // column header label, e.g. "USD" or "KRW" (default "USD")
}

type dataset struct {
	Label       string    `json:"label"`
	Data        []float64 `json:"data"`
	RawData     []float64 `json:"rawData,omitempty"`
	BorderColor string    `json:"borderColor"`
	BorderWidth int       `json:"borderWidth"`
	PointRadius int       `json:"pointRadius"`
	Tension     float64   `json:"tension"`
	Fill        bool      `json:"fill"`
	FinalUSD    float64   `json:"finalUSD"`
	Currency    string    `json:"currency,omitempty"`
}

type tableRow struct {
	Color         string
	Currency      string
	StrategyName  string
	TotalInvested float64
	FinalValue    float64
	ReturnPct     float64
	MDD           float64
	TradeCount    int
	AvgHoldDays   float64
}

type chartSection struct {
	Symbol        string
	Labels        template.JS
	Datasets      template.JS
	Annotations   template.JS
	Rows          []tableRow
	Currency      string
	CurrencyLabel string
}

// mddAnnot describes one vertical annotation line on the chart.
type mddAnnot struct {
	date  string
	color string
	label string
}

// buildAnnotationsJS converts a list of MDD annotations into a Chart.js
// annotation plugin config object (template.JS).
func buildAnnotationsJS(annots []mddAnnot) template.JS {
	// Deduplicate by date, combining labels when multiple strategies share the same date.
	merged := make(map[string]*mddAnnot)
	order := []string{}
	for _, a := range annots {
		if existing, ok := merged[a.date]; ok {
			existing.label += " / " + a.label
		} else {
			cp := a
			merged[a.date] = &cp
			order = append(order, a.date)
		}
	}

	parts := make([]string, 0, len(order))
	for i, date := range order {
		a := merged[date]
		key := fmt.Sprintf("mdd%d", i)
		parts = append(parts, fmt.Sprintf(
			`%q:{type:"line",xMin:%q,xMax:%q,borderColor:%q,borderWidth:1,borderDash:[6,3],`+
				`label:{display:true,content:%q,color:%q,font:{size:10},position:"start",`+
				`backgroundColor:"rgba(0,0,0,0.5)",padding:4}}`,
			key, date, date, a.color, a.label, a.color,
		))
	}
	return template.JS("{" + strings.Join(parts, ",") + "}")
}

// ChartI18n holds all translatable UI strings for the backtest chart.
type ChartI18n struct {
	PageTitle  string
	H1         string
	Strategy   string
	Invested   string
	FinalValue string
	ReturnPct  string
	MDD        string
	Trades     string
	AvgPeriod  string
}

var EnglishChartI18n = ChartI18n{
	PageTitle:  "Strategy Backtest Chart",
	H1:         "Strategy Portfolio Value — Monthly",
	Strategy:   "Strategy",
	Invested:   "Invested",
	FinalValue: "Final Value",
	ReturnPct:  "Return (%)",
	MDD:        "MDD (%)",
	Trades:     "Trades",
	AvgPeriod:  "Avg Period",
}

var KoreanChartI18n = ChartI18n{
	PageTitle:  "전략 백테스트 차트",
	H1:         "전략별 포트폴리오 가치 — 월별",
	Strategy:   "전략",
	Invested:   "투자금",
	FinalValue: "최종 가치",
	ReturnPct:  "수익률 (%)",
	MDD:        "MDD (%)",
	Trades:     "거래 수",
	AvgPeriod:  "평균 보유 기간",
}

// ── Main chart (per-ticker, all strategies) ──────────────────────────────────

func Generate(outputPath string, tickers []TickerChart, start, end time.Time) error {
	return generate(outputPath, tickers, start, end, EnglishChartI18n)
}

func GenerateKorean(outputPath string, tickers []TickerChart, start, end time.Time) error {
	return generate(outputPath, tickers, start, end, KoreanChartI18n)
}

func generate(outputPath string, tickers []TickerChart, start, end time.Time, i18n ChartI18n) error {
	sections := make([]chartSection, 0, len(tickers))
	for _, t := range tickers {
		labels, datasets := buildDatasets(t.Results)
		labelsJSON, _ := json.Marshal(labels)
		datasetsJSON, _ := json.Marshal(datasets)

		// One annotation per unique MDD date across all strategies.
		seen := make(map[string]bool)
		var annots []mddAnnot
		for _, r := range t.Results {
			if r.MDDDate != "" && !seen[r.MDDDate] {
				seen[r.MDDDate] = true
				annots = append(annots, mddAnnot{
					date:  r.MDDDate,
					color: "rgba(255,255,255,0.35)",
					label: "MDD " + r.MDDDate,
				})
			}
		}

		rows := make([]tableRow, len(t.Results))
		for i, r := range t.Results {
			rows[i] = tableRow{
				Color:         palette[i%len(palette)],
				StrategyName:  r.StrategyName,
				TotalInvested: r.TotalInvested,
				FinalValue:    r.FinalValue,
				ReturnPct:     r.ReturnPct(),
				MDD:           r.MDD,
				TradeCount:    r.TradeCount,
				AvgHoldDays:   r.AvgHoldDays,
			}
		}

		cur, curLabel := currencyOr(t.Currency), t.CurrencyLabel
		if curLabel == "" {
			curLabel = "USD"
		}
		sections = append(sections, chartSection{
			Symbol:        t.Symbol,
			Labels:        template.JS(labelsJSON),
			Datasets:      template.JS(datasetsJSON),
			Annotations:   buildAnnotationsJS(annots),
			Rows:          rows,
			Currency:      cur,
			CurrencyLabel: curLabel,
		})
	}

	data := struct {
		Period   string
		Sections []chartSection
		I18n     ChartI18n
	}{
		Period:   fmt.Sprintf("%s → %s", start.Format("2006-01-02"), end.Format("2006-01-02")),
		Sections: sections,
		I18n:     i18n,
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return htmlTemplate.Execute(f, data)
}

func buildDatasets(results []backtester.Result) ([]string, []dataset) {
	if len(results) == 0 {
		return nil, nil
	}
	labels := make([]string, len(results[0].History))
	for i, dp := range results[0].History {
		labels[i] = dp.Date
	}
	labelIndex := make(map[string]int, len(labels))
	for i, l := range labels {
		labelIndex[l] = i
	}
	datasets := make([]dataset, len(results))
	for i, r := range results {
		values := make([]float64, len(labels))
		for _, dp := range r.History {
			if idx, ok := labelIndex[dp.Date]; ok {
				values[idx] = dp.Value
			}
		}
		datasets[i] = dataset{
			Label:       r.StrategyName,
			Data:        values,
			BorderColor: palette[i%len(palette)],
			BorderWidth: 2,
			PointRadius: 0,
			Tension:     0.1,
			Fill:        false,
		}
	}
	return labels, datasets
}

// ── Comparison chart (one strategy, all tickers) ─────────────────────────────

type ComparisonLine struct {
	Label    string
	Color    string
	Currency string // symbol prefix, e.g. "$" or "₩" (default "$")
	Result   backtester.Result
}

func GenerateComparison(outputPath, strategyName string, lines []ComparisonLine, start, end time.Time) error {
	labels := make([]string, len(lines[0].Result.History))
	for i, dp := range lines[0].Result.History {
		labels[i] = dp.Date
	}
	labelIndex := make(map[string]int, len(labels))
	for i, l := range labels {
		labelIndex[l] = i
	}

	datasets := make([]dataset, len(lines))
	var annots []mddAnnot
	for i, l := range lines {
		values := make([]float64, len(labels))
		for _, dp := range l.Result.History {
			if idx, ok := labelIndex[dp.Date]; ok {
				values[idx] = dp.Value
			}
		}
		rawData := make([]float64, len(values))
		copy(rawData, values)

		var baseline float64
		for _, v := range values {
			if v > 0 {
				baseline = v
				break
			}
		}
		if baseline > 0 {
			for j := range values {
				values[j] = values[j] / baseline * 100
			}
		}
		datasets[i] = dataset{
			Label:       l.Label,
			Data:        values,
			RawData:     rawData,
			BorderColor: l.Color,
			BorderWidth: 2,
			PointRadius: 0,
			Tension:     0.1,
			Fill:        false,
			FinalUSD:    l.Result.FinalValue,
			Currency:    currencyOr(l.Currency),
		}
		if l.Result.MDDDate != "" {
			annots = append(annots, mddAnnot{
				date:  l.Result.MDDDate,
				color: l.Color,
				label: l.Label + " MDD",
			})
		}
	}

	labelsJSON, _ := json.Marshal(labels)
	datasetsJSON, _ := json.Marshal(datasets)

	rows := make([]tableRow, len(lines))
	for i, l := range lines {
		rows[i] = tableRow{
			Color:         l.Color,
			Currency:      currencyOr(l.Currency),
			StrategyName:  l.Label,
			TotalInvested: l.Result.TotalInvested,
			FinalValue:    l.Result.FinalValue,
			ReturnPct:     l.Result.ReturnPct(),
			MDD:           l.Result.MDD,
		}
	}

	data := struct {
		Period       string
		StrategyName string
		Labels       template.JS
		Datasets     template.JS
		Annotations  template.JS
		Rows         []tableRow
	}{
		Period:       fmt.Sprintf("%s → %s", start.Format("2006-01-02"), end.Format("2006-01-02")),
		StrategyName: strategyName,
		Labels:       template.JS(labelsJSON),
		Datasets:     template.JS(datasetsJSON),
		Annotations:  buildAnnotationsJS(annots),
		Rows:         rows,
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return comparisonTemplate.Execute(f, data)
}

// ── Shared helpers ────────────────────────────────────────────────────────────

// currencyOr defaults an empty currency symbol to "$".
func currencyOr(cur string) string {
	if cur == "" {
		return "$"
	}
	return cur
}

func formatUSD(v float64) string {
	s := fmt.Sprintf("%.0f", v)
	n := len(s)
	out := make([]byte, 0, n+(n-1)/3+1)
	for i := range s {
		if i > 0 && (n-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, s[i])
	}
	return "$" + string(out)
}

func formatNum(v float64) string {
	s := fmt.Sprintf("%.0f", v)
	n := len(s)
	out := make([]byte, 0, n+(n-1)/3+1)
	for i := range s {
		if i > 0 && (n-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, s[i])
	}
	return string(out)
}

func formatDays(v float64) string {
	if v == 0 {
		return "-"
	}
	return fmt.Sprintf("%.0fd", v)
}

var funcMap = template.FuncMap{
	"printf":     fmt.Sprintf,
	"formatUSD":  formatUSD,
	"formatNum":  formatNum,
	"formatDays": formatDays,
}

// ── HTML templates ────────────────────────────────────────────────────────────

var htmlTemplate = template.Must(template.New("chart").Funcs(funcMap).Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>{{$.I18n.PageTitle}}</title>
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { background: #0f0f1a; color: #e0e0e0; font-family: Arial, sans-serif; padding: 24px; }
  h1 { text-align: center; margin-bottom: 8px; font-size: 1.4rem; color: #ccc; }
  .period { text-align: center; color: #888; font-size: 0.9rem; margin-bottom: 32px; }
  .section { max-width: 1200px; margin: 0 auto 56px; background: #1a1a2e; border-radius: 12px; padding: 24px; }
  .section h2 { font-size: 1.2rem; color: #ddd; margin-bottom: 20px; letter-spacing: 0.05em; }
  table { width: 100%; border-collapse: collapse; margin-top: 24px; font-size: 0.85rem; }
  thead tr { border-bottom: 1px solid #2e2e4e; }
  th { text-align: right; padding: 8px 12px; color: #888; font-weight: normal; }
  th:first-child { text-align: left; }
  td { padding: 8px 12px; text-align: right; border-bottom: 1px solid #1e1e3a; }
  td:first-child { text-align: left; }
  .dot { display: inline-block; width: 10px; height: 10px; border-radius: 50%; margin-right: 8px; flex-shrink: 0; }
  .strategy-cell { display: flex; align-items: center; }
  .pos { color: #2ecc71; }
  .neg { color: #e74c3c; }
  tbody tr { transition: opacity 0.2s; }
  tbody tr.row-hidden { opacity: 0.25; }
</style>
</head>
<body>
<h1>{{$.I18n.H1}}</h1>
<p class="period">{{.Period}}</p>
{{range .Sections}}
<div class="section">
  <h2>{{.Symbol}}</h2>
  <canvas id="chart-{{.Symbol}}"></canvas>
  <table id="table-{{.Symbol}}">
    <thead>
      <tr>
        <th>{{$.I18n.Strategy}}</th>
        <th>{{$.I18n.Invested}} ({{.CurrencyLabel}})</th>
        <th>{{$.I18n.FinalValue}} ({{.CurrencyLabel}})</th>
        <th>{{$.I18n.ReturnPct}}</th>
        <th>{{$.I18n.MDD}}</th>
        <th>{{$.I18n.Trades}}</th>
        <th>{{$.I18n.AvgPeriod}}</th>
      </tr>
    </thead>
    <tbody>
    {{$cur := .Currency}}
    {{range $i, $row := .Rows}}
      <tr data-idx="{{$i}}">
        <td><span class="strategy-cell"><span class="dot" style="background:{{$row.Color}}"></span>{{$row.StrategyName}}</span></td>
        <td>{{$cur}}{{formatNum $row.TotalInvested}}</td>
        <td>{{$cur}}{{formatNum $row.FinalValue}}</td>
        <td class="{{if ge $row.ReturnPct 0.0}}pos{{else}}neg{{end}}">{{printf "%.2f%%" $row.ReturnPct}}</td>
        <td class="neg">{{printf "%.2f%%" $row.MDD}}</td>
        <td>{{$row.TradeCount}}</td>
        <td>{{formatDays $row.AvgHoldDays}}</td>
      </tr>
    {{end}}
    </tbody>
  </table>
</div>
{{end}}
<script src="https://cdn.jsdelivr.net/npm/chart.js@4/dist/chart.umd.min.js"></script>
<script src="https://cdn.jsdelivr.net/npm/chartjs-plugin-annotation@3/dist/chartjs-plugin-annotation.min.js"></script>
<script>
{{range .Sections}}
new Chart(document.getElementById('chart-{{.Symbol}}'), {
  type: 'line',
  data: { labels: {{.Labels}}, datasets: {{.Datasets}} },
  options: {
    responsive: true,
    interaction: { mode: 'index', intersect: false },
    plugins: {
      legend: {
        labels: { color: '#ccc', boxWidth: 14, font: { size: 11 } },
        onClick(e, item, legend) {
          Chart.defaults.plugins.legend.onClick.call(this, e, item, legend);
          const idx = item.datasetIndex;
          const symbol = legend.chart.canvas.id.replace('chart-', '');
          const row = document.querySelector('#table-' + symbol + ' tbody tr[data-idx="' + idx + '"]');
          if (row) row.classList.toggle('row-hidden', !legend.chart.isDatasetVisible(idx));
        }
      },
      tooltip: {
        callbacks: {
          label: ctx => ctx.dataset.label + ': {{.Currency}}' + ctx.parsed.y.toLocaleString(undefined, {maximumFractionDigits: 0})
        }
      },
      annotation: { annotations: {{.Annotations}} }
    },
    scales: {
      x: { ticks: { color: '#888', maxTicksLimit: 20, maxRotation: 45 }, grid: { color: '#2a2a3e' } },
      y: {
        type: 'linear',
        ticks: { color: '#888', callback: v => '{{.Currency}}' + v.toLocaleString() },
        grid: { color: '#2a2a3e' }
      }
    }
  }
});
{{end}}
</script>
</body>
</html>
`))

var comparisonTemplate = template.Must(template.New("comparison").Funcs(funcMap).Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>{{.StrategyName}} — Ticker Comparison</title>
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { background: #0f0f1a; color: #e0e0e0; font-family: Arial, sans-serif; padding: 24px; }
  h1 { text-align: center; margin-bottom: 8px; font-size: 1.4rem; color: #ccc; }
  .period { text-align: center; color: #888; font-size: 0.9rem; margin-bottom: 32px; }
  .section { max-width: 1200px; margin: 0 auto; background: #1a1a2e; border-radius: 12px; padding: 24px; }
  table { width: 100%; border-collapse: collapse; margin-top: 24px; font-size: 0.85rem; }
  thead tr { border-bottom: 1px solid #2e2e4e; }
  th { text-align: right; padding: 8px 12px; color: #888; font-weight: normal; }
  th:first-child { text-align: left; }
  td { padding: 8px 12px; text-align: right; border-bottom: 1px solid #1e1e3a; }
  td:first-child { text-align: left; }
  .dot { display: inline-block; width: 10px; height: 10px; border-radius: 50%; margin-right: 8px; }
  .ticker-cell { display: flex; align-items: center; }
  .pos { color: #2ecc71; }
  .neg { color: #e74c3c; }
</style>
</head>
<body>
<h1>{{.StrategyName}}</h1>
<p class="period">{{.Period}}</p>
<div class="section">
  <canvas id="chart"></canvas>
  <table>
    <thead>
      <tr>
        <th>Ticker</th>
        <th>Invested</th>
        <th>Final Value</th>
        <th>Return (%)</th>
        <th>MDD (%)</th>
      </tr>
    </thead>
    <tbody>
    {{range .Rows}}
      <tr>
        <td><span class="ticker-cell"><span class="dot" style="background:{{.Color}}"></span>{{.StrategyName}}</span></td>
        <td>{{.Currency}}{{formatNum .TotalInvested}}</td>
        <td>{{.Currency}}{{formatNum .FinalValue}}</td>
        <td class="{{if ge .ReturnPct 0.0}}pos{{else}}neg{{end}}">{{printf "%.2f%%" .ReturnPct}}</td>
        <td class="neg">{{printf "%.2f%%" .MDD}}</td>
      </tr>
    {{end}}
    </tbody>
  </table>
</div>
<script src="https://cdn.jsdelivr.net/npm/chart.js@4/dist/chart.umd.min.js"></script>
<script src="https://cdn.jsdelivr.net/npm/chartjs-plugin-annotation@3/dist/chartjs-plugin-annotation.min.js"></script>
<script src="https://cdn.jsdelivr.net/npm/chartjs-plugin-datalabels@2/dist/chartjs-plugin-datalabels.min.js"></script>
<script>
Chart.register(ChartDataLabels);
new Chart(document.getElementById('chart'), {
  type: 'line',
  data: { labels: {{.Labels}}, datasets: {{.Datasets}} },
  options: {
    responsive: true,
    layout: { padding: { right: 110 } },
    interaction: { mode: 'index', intersect: false },
    plugins: {
      legend: { labels: { color: '#ccc', boxWidth: 14, font: { size: 12 } } },
      tooltip: {
        callbacks: {
          label: ctx => {
            const raw = ctx.dataset.rawData?.[ctx.dataIndex];
            const cur = ctx.dataset.currency || '$';
            return ctx.dataset.label + ': ' + (raw != null ? cur + raw.toLocaleString(undefined, {maximumFractionDigits: 0}) : '');
          }
        }
      },
      datalabels: {
        display: ctx => ctx.dataIndex === ctx.dataset.data.length - 1,
        anchor: 'end', align: 'right',
        color: ctx => ctx.dataset.borderColor,
        font: { size: 11, weight: 'bold' },
        formatter: (_, ctx) => (ctx.dataset.currency || '$') + ctx.dataset.finalUSD.toLocaleString(undefined, {maximumFractionDigits: 0})
      },
      annotation: { annotations: {{.Annotations}} }
    },
    scales: {
      x: { ticks: { color: '#888', maxTicksLimit: 20, maxRotation: 45 }, grid: { color: '#2a2a3e' } },
      y: {
        type: 'linear',
        ticks: { color: '#888', callback: v => v.toLocaleString() + '%' },
        grid: { color: '#2a2a3e' }
      }
    }
  }
});
</script>
</body>
</html>
`))
