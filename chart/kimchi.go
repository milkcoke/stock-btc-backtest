package chart

import (
	"encoding/json"
	"html/template"
	"math"
	"os"
	"sort"

	"stock-btc-backtest/loader"
)

type kimchiData struct {
	Labels     template.JS
	USDData    template.JS
	USDTData   template.JS
	WonData    template.JS
	PctData    template.JS
	AvgWon     float64
	AvgPct     float64
	MedianWon  float64
	MedianPct  float64
	MaxWon     float64
	MaxPct     float64
	MinWon     float64
	MinPct     float64
	MaxWonDate string
	MinWonDate string
}

func GenerateKimchi(outputPath string, records []loader.KimchiRecord) error {
	labels := make([]string, len(records))
	usdData := make([]float64, len(records))
	usdtData := make([]float64, len(records))
	wonData := make([]float64, len(records))
	pctData := make([]float64, len(records))

	maxWon, minWon := math.Inf(-1), math.Inf(1)
	maxPct, minPct := math.Inf(-1), math.Inf(1)
	var sumWon, sumPct float64
	maxWonDate, minWonDate := "", ""
	rawWon := make([]float64, len(records))
	rawPct := make([]float64, len(records))

	for i, r := range records {
		labels[i] = r.Date
		usdData[i] = math.Round(r.USDKRW*100) / 100
		usdtData[i] = math.Round(r.USDTKRW*100) / 100
		wonData[i] = math.Round(r.PremiumWon*100) / 100
		pctData[i] = math.Round(r.PremiumPct*10000) / 10000
		rawWon[i] = r.PremiumWon
		rawPct[i] = r.PremiumPct
		sumWon += r.PremiumWon
		sumPct += r.PremiumPct
		if r.PremiumWon > maxWon {
			maxWon = r.PremiumWon
			maxWonDate = r.Date
		}
		if r.PremiumWon < minWon {
			minWon = r.PremiumWon
			minWonDate = r.Date
		}
		if r.PremiumPct > maxPct {
			maxPct = r.PremiumPct
		}
		if r.PremiumPct < minPct {
			minPct = r.PremiumPct
		}
	}

	n := float64(len(records))
	sort.Float64s(rawWon)
	sort.Float64s(rawPct)
	labelsJSON, _ := json.Marshal(labels)
	usdJSON, _ := json.Marshal(usdData)
	usdtJSON, _ := json.Marshal(usdtData)
	wonJSON, _ := json.Marshal(wonData)
	pctJSON, _ := json.Marshal(pctData)

	d := kimchiData{
		Labels:     template.JS(labelsJSON),
		USDData:    template.JS(usdJSON),
		USDTData:   template.JS(usdtJSON),
		WonData:    template.JS(wonJSON),
		PctData:    template.JS(pctJSON),
		AvgWon:     sumWon / n,
		AvgPct:     sumPct / n,
		MedianWon:  median(rawWon),
		MedianPct:  median(rawPct),
		MaxWon:     maxWon,
		MaxPct:     maxPct,
		MinWon:     minWon,
		MinPct:     minPct,
		MaxWonDate: maxWonDate,
		MinWonDate: minWonDate,
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return kimchiTemplate.Execute(f, d)
}

func median(sorted []float64) float64 {
	n := len(sorted)
	if n == 0 {
		return 0
	}
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}

var kimchiTemplate = template.Must(template.New("kimchi").Funcs(funcMap).Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>Kimchi Premium</title>
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { background: #0f0f1a; color: #e0e0e0; font-family: Arial, sans-serif; padding: 24px; }
  h1 { text-align: center; margin-bottom: 8px; font-size: 1.4rem; color: #ccc; }
  .subtitle { text-align: center; color: #888; font-size: 0.85rem; margin-bottom: 32px; }
  .section { max-width: 1200px; margin: 0 auto 32px; background: #1a1a2e; border-radius: 12px; padding: 24px; }
  .section h2 { font-size: 1rem; color: #aaa; margin-bottom: 16px; letter-spacing: 0.04em; }
  .stats { display: flex; gap: 16px; margin-top: 24px; flex-wrap: wrap; }
  .stat-card { flex: 1; min-width: 140px; background: #12122a; border-radius: 8px; padding: 14px 18px; }
  .stat-card .label { font-size: 0.75rem; color: #888; margin-bottom: 6px; }
  .stat-card .val { font-size: 1.05rem; font-weight: bold; }
  .stat-card .sub { font-size: 0.78rem; color: #aaa; margin-top: 4px; }
  .pos { color: #2ecc71; }
  .neg { color: #e74c3c; }
  .neu { color: #f1c40f; }
</style>
</head>
<body>
<h1>Kimchi Premium — Daily</h1>
<p class="subtitle">USDT/KRW (Upbit) − USD/KRW (Yahoo Finance)</p>

<div class="section">
  <h2>Kimchi Premium</h2>
  <canvas id="chart-premium"></canvas>
  <div class="stats">
    <div class="stat-card">
      <div class="label">Average Premium</div>
      <div class="val neu">{{printf "%.2f" .AvgWon}} ₩</div>
      <div class="sub">{{printf "%.4f" .AvgPct}}%</div>
    </div>
    <div class="stat-card">
      <div class="label">Median Premium (p50)</div>
      <div class="val neu">{{printf "%.2f" .MedianWon}} ₩</div>
      <div class="sub">{{printf "%.4f" .MedianPct}}%</div>
    </div>
    <div class="stat-card">
      <div class="label">Max Premium</div>
      <div class="val pos">{{printf "%.2f" .MaxWon}} ₩</div>
      <div class="sub">{{printf "%.4f" .MaxPct}}% · {{.MaxWonDate}}</div>
    </div>
    <div class="stat-card">
      <div class="label">Min Premium</div>
      <div class="val neg">{{printf "%.2f" .MinWon}} ₩</div>
      <div class="sub">{{printf "%.4f" .MinPct}}% · {{.MinWonDate}}</div>
    </div>
  </div>
</div>

<div class="section">
  <h2>Exchange Rate</h2>
  <canvas id="chart-rate"></canvas>
</div>

<script src="https://cdn.jsdelivr.net/npm/chart.js@4/dist/chart.umd.min.js"></script>
<script src="https://cdn.jsdelivr.net/npm/chartjs-plugin-annotation@3/dist/chartjs-plugin-annotation.min.js"></script>
<script>
const labels = {{.Labels}};
const sharedX = {
  ticks: { color: '#888', maxTicksLimit: 24, maxRotation: 45 },
  grid: { color: '#2a2a3e' }
};

// ── Chart 1: USD/KRW and USDT/KRW price ──────────────────────────────────────
new Chart(document.getElementById('chart-rate'), {
  type: 'line',
  data: {
    labels,
    datasets: [
      {
        label: 'USD/KRW',
        data: {{.USDData}},
        borderColor: '#2ecc71',
        borderWidth: 1.5,
        pointRadius: 0,
        tension: 0.2,
        fill: false,
      },
      {
        label: 'USDT/KRW',
        data: {{.USDTData}},
        borderColor: '#e74c3c',
        borderWidth: 1.5,
        pointRadius: 0,
        tension: 0.2,
        fill: false,
      }
    ]
  },
  options: {
    responsive: true,
    interaction: { mode: 'index', intersect: false },
    plugins: {
      legend: { labels: { color: '#ccc', boxWidth: 14, font: { size: 12 } } },
      tooltip: {
        callbacks: {
          label: ctx => ctx.dataset.label + ': ₩' + ctx.parsed.y.toLocaleString(undefined, {maximumFractionDigits: 2})
        }
      }
    },
    scales: {
      x: sharedX,
      y: {
        type: 'linear',
        position: 'left',
        ticks: { color: '#888', callback: v => '₩' + v.toLocaleString() },
        grid: { color: '#2a2a3e' },
        title: { display: true, text: 'Exchange Rate (₩)', color: '#888', font: { size: 11 } }
      }
    }
  }
});

// ── Chart 2: Kimchi premium ₩ and % ──────────────────────────────────────────
new Chart(document.getElementById('chart-premium'), {
  type: 'line',
  data: {
    labels,
    datasets: [
      {
        label: 'Premium (₩)',
        data: {{.WonData}},
        yAxisID: 'y',
        borderColor: '#f1c40f',
        backgroundColor: 'rgba(241,196,15,0.08)',
        borderWidth: 1.5,
        pointRadius: 0,
        tension: 0.2,
        fill: 'origin',
      },
      {
        label: 'Premium (%)',
        data: {{.PctData}},
        yAxisID: 'y1',
        borderColor: '#3498db',
        borderWidth: 1.5,
        pointRadius: 0,
        tension: 0.2,
        fill: false,
      }
    ]
  },
  options: {
    responsive: true,
    interaction: { mode: 'index', intersect: false },
    plugins: {
      legend: { labels: { color: '#ccc', boxWidth: 14, font: { size: 12 } } },
      tooltip: {
        callbacks: {
          label: ctx => ctx.datasetIndex === 0
            ? ctx.dataset.label + ': ₩' + ctx.parsed.y.toLocaleString(undefined, {maximumFractionDigits: 2})
            : ctx.dataset.label + ': ' + ctx.parsed.y.toFixed(4) + '%'
        }
      },
      annotation: {
        annotations: {
          zero: {
            type: 'line', yMin: 0, yMax: 0, yScaleID: 'y',
            borderColor: 'rgba(255,255,255,0.25)', borderWidth: 1, borderDash: [4, 4]
          }
        }
      }
    },
    scales: {
      x: sharedX,
      y: {
        type: 'linear',
        position: 'left',
        ticks: { color: '#f1c40f', callback: v => '₩' + v.toLocaleString() },
        grid: { color: '#2a2a3e' },
        title: { display: true, text: 'Premium (₩)', color: '#f1c40f', font: { size: 11 } }
      },
      y1: {
        type: 'linear',
        position: 'right',
        ticks: { color: '#3498db', callback: v => v.toFixed(2) + '%' },
        grid: { drawOnChartArea: false },
        title: { display: true, text: 'Premium (%)', color: '#3498db', font: { size: 11 } }
      }
    }
  }
});
</script>
</body>
</html>
`))
