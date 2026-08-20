package chart

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"strings"

	"stock-btc-backtest/eventstudy"
	"stock-btc-backtest/valuation"
)

// entryView is the flattened shape the entry template renders. Building it here
// keeps date formatting and JSON encoding out of both the template and the
// study.
type entryView struct {
	Symbol   string
	Period   string
	Rule     string
	MDDNote  string
	Currency string

	Labels     template.JS
	Prices     template.JS
	SMA        template.JS
	MALabel    string
	Markers    template.JS
	Annotation template.JS

	Targets  []targetRow
	Entries  []entryRow
	Censored string

	HasValuation bool
	ValNote      string
	ValLabels    template.JS
	Panels       []valPanel
}

// valPanel is one ratio's block: heading, stat strip, chart and a folded table.
type valPanel struct {
	Key         string // canvas id suffix
	Name        string
	Note        string
	Color       string
	Series      template.JS
	Median      float64
	WindowLabel string // "10년", used in the stat captions
	Stats       []statCell
	Rows        []valYearRow
	Total       valYearRow
}

type statCell struct {
	Key, Value, Detail string
}

type valYearRow struct {
	Label                 string
	N                     int
	Median, Avg, Min, Max string
	Partial               bool
}

type targetRow struct {
	Target                 float64
	Achieved, Total        int
	Mean, Median, Min, Max string
}

type entryRow struct {
	BuyDate, Buy, VsPeak, VsMA, RSI, LowDate, Low string
	Targets                                       []string
}

// GenerateEntry writes the event study report. A nil valuation summary drops the
// PER/PBR canvases rather than drawing empty ones.
func GenerateEntry(outputPath string, res eventstudy.Result, val *valuation.Summary) error {
	v := entryView{
		Symbol: res.Symbol,
		Period: fmt.Sprintf("%s → %s · %d년 · 거래일 %d일",
			res.From.Format("2006-01-02"), res.To.Format("2006-01-02"),
			res.Years, len(res.Dates)),
		Rule:    res.Rule,
		MALabel: fmt.Sprintf("MA%d", maWindowOf(res)),
	}

	if res.MDDDerived {
		v.MDDNote = fmt.Sprintf("MDD %.2f%% — 완전 연도 %d개의 평균 (중위값 %.2f%%)",
			res.MDDThreshold*100, res.MDDSummary.N, res.MDDSummary.Median)
	} else {
		v.MDDNote = fmt.Sprintf("MDD %.2f%% (직접 지정)", res.MDDThreshold*100)
	}

	labels := make([]string, len(res.Dates))
	for i, d := range res.Dates {
		labels[i] = d.Format("2006-01-02")
	}
	v.Labels = jsonJS(labels)
	v.Prices = jsonJS(res.Prices)
	v.SMA = jsonJS(nanToNull(res.SMA))

	markers := make([]map[string]any, 0, len(res.Entries))
	annots := make([]string, 0, len(res.Entries))
	for _, e := range res.Entries {
		day := e.BuyDate.Format("2006-01-02")
		markers = append(markers, map[string]any{"x": day, "y": e.Buy})
		annots = append(annots, fmt.Sprintf(
			`{type:'line',scaleID:'x',value:%q,borderColor:'#e74c3c',borderWidth:1,borderDash:[4,4]}`, day))
	}
	v.Markers = jsonJS(markers)
	v.Annotation = template.JS("[" + strings.Join(annots, ",") + "]")

	for _, t := range res.Targets {
		row := targetRow{Target: t.Target, Achieved: t.Achieved, Total: t.Total,
			Mean: "—", Median: "—", Min: "—", Max: "—"}
		if t.Summary.N > 0 {
			row.Mean = fmt.Sprintf("%.0f일", t.Summary.Mean)
			row.Median = fmt.Sprintf("%.0f일", t.Summary.Median)
			row.Min = fmt.Sprintf("%.0f일", t.Summary.Min)
			row.Max = fmt.Sprintf("%.0f일", t.Summary.Max)
		}
		v.Targets = append(v.Targets, row)
	}

	var censored []string
	for _, e := range res.Entries {
		row := entryRow{
			BuyDate: e.BuyDate.Format("2006-01-02"), Buy: formatNum(e.Buy),
			VsPeak: fmt.Sprintf("%.1f%%", e.VsPeakPct), VsMA: fmt.Sprintf("%.1f%%", e.VsMAPct),
			RSI:     fmt.Sprintf("%.1f", e.RSI),
			LowDate: e.LowDate.Format("2006-01-02"), Low: fmt.Sprintf("%.1f%%", e.LowPct),
		}
		for _, o := range e.Outcomes {
			if o.Achieved {
				row.Targets = append(row.Targets, fmt.Sprintf("%d일", o.Days))
			} else {
				row.Targets = append(row.Targets, fmt.Sprintf("미달성 (%d일째)", o.Days))
				censored = append(censored, fmt.Sprintf("%s +%.0f%%", row.BuyDate, o.Target))
			}
		}
		v.Entries = append(v.Entries, row)
	}
	if len(censored) > 0 {
		v.Censored = fmt.Sprintf("미달성 %d건은 통계에서 제외했다 — 아직 시간이 안 된 미완결 관측치이지 느린 사례가 아니다: %s",
			len(censored), strings.Join(censored, ", "))
	}

	if val != nil && val.PER.N > 0 {
		v.HasValuation = true
		note := fmt.Sprintf("%s → %s · 거래일 %d일 · 실제 확보 %.1f년",
			val.From.Format("2006-01-02"), val.To.Format("2006-01-02"), val.PER.N, val.CoveredYears)
		if val.Short() {
			note += fmt.Sprintf(" — 요청 %d년보다 짧다", val.RequestedYrs)
		}
		v.ValNote = note

		vl := make([]string, len(val.Records))
		for i, r := range val.Records {
			vl[i] = r.Date.Format("2006-01-02")
		}
		v.ValLabels = jsonJS(vl)
		v.Panels = valuationPanels(val)
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return entryTemplate.Execute(f, v)
}

// maWindowOf recovers the moving-average window from the rule description so the
// legend matches the rule without widening Result for a label.
// valuationPanels builds one block per ratio. The captions carry the dates of
// the extremes because "max 19.08" only becomes information once you know it
// was January 2018.
func valuationPanels(val *valuation.Summary) []valPanel {
	window := fmt.Sprintf("%.0f년", val.CoveredYears)
	specs := []struct {
		key, note, color string
		metric           valuation.Metric
		pick             func(valuation.YearSummary) statsView
	}{
		{"per", "주가수익비율 — 종가 ÷ 주당순이익(EPS). 낮을수록 이익 대비 저평가.", "--series-per",
			val.PER, func(y valuation.YearSummary) statsView {
				return statsView{y.PER.N, y.PER.Median, y.PER.Mean, y.PER.Min, y.PER.Max}
			}},
		{"pbr", "주가순자산비율 — 종가 ÷ 주당순자산(BPS). 1.0 미만이면 장부가 밑에서 거래.", "--series-pbr",
			val.PBR, func(y valuation.YearSummary) statsView {
				return statsView{y.PBR.N, y.PBR.Median, y.PBR.Mean, y.PBR.Min, y.PBR.Max}
			}},
	}

	panels := make([]valPanel, 0, len(specs))
	for _, spec := range specs {
		m := spec.metric
		series := make([]float64, len(val.Records))
		for i, r := range val.Records {
			if spec.key == "per" {
				series[i] = r.PER
			} else {
				series[i] = r.PBR
			}
		}

		p := valPanel{
			Key: spec.key, Name: m.Name, Note: spec.note, Color: spec.color,
			Series: jsonJS(series), Median: m.Median, WindowLabel: window,
			Stats: []statCell{
				{"median", fmt.Sprintf("%.2f", m.Median), window + " 중앙값"},
				{"avg", fmt.Sprintf("%.2f", m.Mean), window + " 평균"},
				{"min", fmt.Sprintf("%.2f", m.Min), m.MinDate.Format("2006-01-02")},
				{"max", fmt.Sprintf("%.2f", m.Max), m.MaxDate.Format("2006-01-02")},
				{"현재", fmt.Sprintf("%.2f", m.Last), m.LastDate.Format("2006-01-02")},
			},
		}
		for _, y := range val.Yearly {
			s := spec.pick(y)
			p.Rows = append(p.Rows, valYearRow{
				Label: fmt.Sprintf("%d", y.Year), N: s.n,
				Median: fmt.Sprintf("%.2f", s.median), Avg: fmt.Sprintf("%.2f", s.avg),
				Min: fmt.Sprintf("%.2f", s.min), Max: fmt.Sprintf("%.2f", s.max),
				Partial: y.Year == val.From.Year() || y.Year == val.To.Year(),
			})
		}
		p.Total = valYearRow{
			Label: window + " 전체", N: m.N,
			Median: fmt.Sprintf("%.2f", m.Median), Avg: fmt.Sprintf("%.2f", m.Mean),
			Min: fmt.Sprintf("%.2f", m.Min), Max: fmt.Sprintf("%.2f", m.Max),
		}
		panels = append(panels, p)
	}
	return panels
}

type statsView struct {
	n                     int
	median, avg, min, max float64
}

func maWindowOf(res eventstudy.Result) int {
	var w int
	if _, err := fmt.Sscanf(afterMA(res.Rule), "%d", &w); err == nil && w > 0 {
		return w
	}
	return 200
}

func afterMA(rule string) string {
	if i := strings.Index(rule, "MA"); i >= 0 {
		return rule[i+2:]
	}
	return ""
}

// nanToNull turns warm-up NaNs into nulls so Chart.js leaves a gap instead of
// refusing to draw the series.
func nanToNull(values []float64) []any {
	out := make([]any, len(values))
	for i, v := range values {
		if v != v {
			out[i] = nil
		} else {
			out[i] = v
		}
	}
	return out
}

func jsonJS(v any) template.JS {
	b, err := json.Marshal(v)
	if err != nil {
		return template.JS("[]")
	}
	return template.JS(b)
}
