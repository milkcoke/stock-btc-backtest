package chart

import "html/template"

var entryTemplate = template.Must(template.New("entry").Funcs(funcMap).Parse(`<!DOCTYPE html>
<html lang="ko">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Symbol}} 가치 진입 분석</title>
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { background: #0f0f1a; color: #e0e0e0; font-family: -apple-system, BlinkMacSystemFont,
         "Apple SD Gothic Neo", Arial, sans-serif; padding: 24px; }
  h1 { text-align: center; margin-bottom: 8px; font-size: 1.4rem; color: #ccc; }
  .period { text-align: center; color: #888; font-size: 0.9rem; }
  .rule { text-align: center; color: #6fa8dc; font-size: 0.95rem; margin: 10px 0 4px; }
  .note { text-align: center; color: #8a8a9a; font-size: 0.82rem; margin-bottom: 28px; }
  .section { max-width: 1200px; margin: 0 auto 48px; background: #1a1a2e; border-radius: 12px; padding: 24px; }
  .section h2 { font-size: 1.1rem; color: #ddd; margin-bottom: 18px; letter-spacing: 0.05em; }
  table { width: 100%; border-collapse: collapse; margin-top: 20px; font-size: 0.85rem; }
  thead tr { border-bottom: 1px solid #2e2e4e; }
  th { text-align: right; padding: 8px 12px; color: #888; font-weight: normal; }
  th:first-child { text-align: left; }
  td { padding: 8px 12px; text-align: right; border-bottom: 1px solid #1e1e3a; }
  td:first-child { text-align: left; }
  .warn { color: #e0a458; font-size: 0.82rem; margin-top: 16px; line-height: 1.6; }
  .disclaimer { max-width: 1200px; margin: 0 auto; color: #6c6c7c; font-size: 0.8rem;
                text-align: center; line-height: 1.7; }
  :root { --series-per: #3987e5; --series-pbr: #d95926; }
  .phead { display: flex; align-items: center; gap: 10px; margin-bottom: 6px; }
  .swatch { width: 11px; height: 11px; border-radius: 50%; flex-shrink: 0; }
  .ptitle { font-size: 1.15rem; color: #eee; letter-spacing: 0.02em; margin: 0; }
  .pnote { color: #8a8a9a; font-size: 0.85rem; margin-bottom: 20px; }
  .stats { display: flex; flex-wrap: wrap; gap: 34px; margin-bottom: 22px; }
  .stat .k { color: #7c7c8c; font-size: 0.72rem; letter-spacing: 0.08em; text-transform: uppercase; }
  .stat .v { font-size: 1.7rem; font-weight: 600; color: #f2f2f2; line-height: 1.25;
             font-variant-numeric: tabular-nums; }
  .stat .d { color: #6c6c7c; font-size: 0.75rem; }
  details { margin-top: 20px; }
  details summary { cursor: pointer; color: #8fb8e8; font-size: 0.85rem; padding: 6px 0;
                    list-style: none; user-select: none; }
  details summary::-webkit-details-marker { display: none; }
  details summary::before { content: '▶ '; font-size: 0.7rem; }
  details[open] summary::before { content: '▼ '; }
  details table { margin-top: 8px; }
  tfoot td { border-top: 1px solid #2e2e4e; border-bottom: none; color: #bbb; font-weight: 600; }
</style>
</head>
<body>
<h1>{{.Symbol}} — 가치 진입 분석</h1>
<p class="period">{{.Period}}</p>
<p class="rule">{{.Rule}}</p>
<p class="note">{{.MDDNote}}</p>

<div class="section">
  <h2>주가와 진입 시점</h2>
  <canvas id="priceChart"></canvas>
</div>

<div class="section">
  <h2>목표 수익률 도달 기간</h2>
  <table>
    <thead><tr><th>목표</th><th>달성</th><th>평균</th><th>중위값</th><th>최소</th><th>최대</th></tr></thead>
    <tbody>
    {{range .Targets}}
      <tr>
        <td>+{{printf "%.0f" .Target}}%</td>
        <td>{{.Achieved}}/{{.Total}}</td>
        <td>{{.Mean}}</td><td>{{.Median}}</td><td>{{.Min}}</td><td>{{.Max}}</td>
      </tr>
    {{end}}
    </tbody>
  </table>
  {{if .Censored}}<p class="warn">{{.Censored}}</p>{{end}}
</div>

<div class="section">
  <h2>진입 사례</h2>
  <table>
    <thead><tr>
      <th>매수일</th><th>매수가</th><th>고점 대비</th><th>MA 대비</th><th>RSI</th>
      <th>이후 최저일</th><th>최저</th>
      {{range .Targets}}<th>+{{printf "%.0f" .Target}}%</th>{{end}}
    </tr></thead>
    <tbody>
    {{range .Entries}}
      <tr>
        <td>{{.BuyDate}}</td><td>{{.Buy}}</td><td>{{.VsPeak}}</td><td>{{.VsMA}}</td>
        <td>{{.RSI}}</td><td>{{.LowDate}}</td><td>{{.Low}}</td>
        {{range .Targets}}<td>{{.}}</td>{{end}}
      </tr>
    {{end}}
    </tbody>
  </table>
</div>

{{if .HasValuation}}
{{range .Panels}}
<div class="section">
  <div class="phead">
    <span class="swatch" style="background: var({{.Color}})"></span>
    <h2 class="ptitle">{{.Name}}</h2>
  </div>
  <p class="pnote">{{.Note}}</p>
  <div class="stats">
  {{range .Stats}}
    <div class="stat">
      <div class="k">{{.Key}}</div><div class="v">{{.Value}}</div><div class="d">{{.Detail}}</div>
    </div>
  {{end}}
  </div>
  <canvas id="chart-{{.Key}}"></canvas>
  <details>
    <summary>연도별 {{.Name}} 표 (median · avg · min · max)</summary>
    <table>
      <thead><tr><th>연도</th><th>거래일</th><th>median</th><th>avg</th><th>min</th><th>max</th></tr></thead>
      <tbody>
      {{range .Rows}}
        <tr><td>{{.Label}}{{if .Partial}} *{{end}}</td><td>{{.N}}</td>
            <td>{{.Median}}</td><td>{{.Avg}}</td><td>{{.Min}}</td><td>{{.Max}}</td></tr>
      {{end}}
      </tbody>
      <tfoot><tr><td>{{.Total.Label}}</td><td>{{.Total.N}}</td>
        <td>{{.Total.Median}}</td><td>{{.Total.Avg}}</td>
        <td>{{.Total.Min}}</td><td>{{.Total.Max}}</td></tr></tfoot>
    </table>
    <p class="pnote" style="margin:10px 0 0">* 창의 첫 해와 마지막 해는 부분 연도라 다른 해와 직접 비교할 수 없다.</p>
  </details>
</div>
{{end}}
<p class="disclaimer" style="margin-bottom:32px">{{.ValNote}}</p>
{{else}}
<div class="section">
  <h2>PER / PBR</h2>
  <p class="warn">밸류에이션 데이터가 없어 생략했다. 미국 상장사는 SEC 공시에서 자동 산출되고,
     그 외 종목은 data/{ticker}_per_pbr.csv 가 필요하다.</p>
</div>
{{end}}

<p class="disclaimer">
종가 기준. 배당 재투자·세금·수수료 미반영. 하락 국면당 최초 1회 전액 매수 가정이라 실제보다 불리한 체결을 가정한 값이다.<br>
표본이 10건 내외라 평균·중위값은 추정치가 아니라 기록이다. 과거 사례 집계이며 투자 판단의 근거가 아니다.
</p>

<script src="https://cdn.jsdelivr.net/npm/chart.js@4/dist/chart.umd.min.js"></script>
<script src="https://cdn.jsdelivr.net/npm/chartjs-plugin-annotation@3/dist/chartjs-plugin-annotation.min.js"></script>
<script>
const commonScales = {
  x: { ticks: { color: '#777', maxTicksLimit: 14 }, grid: { color: '#22223a' } }
};

new Chart(document.getElementById('priceChart'), {
  type: 'line',
  data: {
    labels: {{.Labels}},
    datasets: [
      { label: '{{.Symbol}}', data: {{.Prices}}, borderColor: '#3987e5',
        borderWidth: 1.5, pointRadius: 0, tension: 0 },
      { label: '{{.MALabel}}', data: {{.SMA}}, borderColor: '#8d8c82',
        borderWidth: 1, pointRadius: 0, tension: 0, spanGaps: false },
      { label: '진입', data: {{.Markers}}, type: 'scatter', showLine: false,
        borderColor: '#e74c3c', backgroundColor: 'transparent',
        pointRadius: 6, pointBorderWidth: 2 }
    ]
  },
  options: {
    responsive: true, interaction: { mode: 'index', intersect: false },
    plugins: {
      legend: { labels: { color: '#aaa' } },
      annotation: { annotations: {{.Annotation}} }
    },
    scales: {
      x: commonScales.x,
      // 20년 주가는 저점 대비 고점이 수십 배라 선형 축에서는 초기 구간이 바닥에 눌린다.
      y: { type: 'logarithmic', ticks: { color: '#777' }, grid: { color: '#22223a' } }
    }
  }
});

{{if .HasValuation}}
const valLabels = {{.ValLabels}};
function valuationChart(id, label, data, color, median) {
  new Chart(document.getElementById(id), {
    type: 'line',
    data: {
      labels: valLabels,
      datasets: [
        { label: label, data: data, borderColor: color, borderWidth: 1.5, pointRadius: 0, tension: 0 },
        { label: '중앙값 ' + median.toFixed(2), data: data.map(() => median),
          borderColor: '#8d8c82', borderWidth: 1, borderDash: [5, 4], pointRadius: 0 }
      ]
    },
    options: {
      responsive: true, interaction: { mode: 'index', intersect: false },
      plugins: { legend: { labels: { color: '#aaa' } } },
      scales: { x: commonScales.x, y: { ticks: { color: '#777' }, grid: { color: '#22223a' } } }
    }
  });
}
{{range .Panels}}
valuationChart('chart-{{.Key}}', '{{.Name}}', {{.Series}},
  getComputedStyle(document.documentElement).getPropertyValue('{{.Color}}').trim(), {{.Median}});
{{end}}
{{end}}
</script>
</body>
</html>
`))
