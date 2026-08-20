package reporter

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"stock-btc-backtest/eventstudy"
	"stock-btc-backtest/valuation"
)

// EntryReporter prints an event study. It formats; it never computes.
type EntryReporter struct {
	Currency string // symbol prefix, e.g. "$" or "₩" (default "$")
}

func (r EntryReporter) Print(res eventstudy.Result, val *valuation.Summary) {
	cur := r.Currency
	if cur == "" {
		cur = "$"
	}

	fmt.Printf("\n[ %s ]  %s → %s  (%d년, 거래일 %d일, 가격열 %s)\n",
		res.Symbol, res.From.Format("2006-01-02"), res.To.Format("2006-01-02"),
		res.Years, len(res.Dates), res.PriceColumn)
	fmt.Printf("진입 조건: %s\n", res.Rule)

	if res.MDDDerived {
		fmt.Printf("낙폭 임계값 %.2f%% = 완전 연도 %d개의 평균 MDD (중위값 %.2f%%, 최악 %.2f%%)\n",
			res.MDDThreshold*100, res.MDDSummary.N, res.MDDSummary.Median, res.MDDSummary.Min)
		fmt.Println("  ※ 창 전체 평균이라 과거 진입 판정에 이후 데이터가 쓰인다 (in-sample 참고 지표)")
	} else {
		fmt.Printf("낙폭 임계값 %.2f%% (직접 지정)\n", res.MDDThreshold*100)
	}

	r.printYearlyMDD(res)
	r.printEntries(res, cur)
	r.printTargets(res)
	r.printValuation(val)

	fmt.Println("\n종가 기준. 배당 재투자·세금·수수료 미반영. 하락 국면당 최초 1회 전액 매수 가정이라")
	fmt.Println("실제보다 불리한 체결을 가정한 값이다. 과거 사례 집계이며 투자 판단의 근거가 아니다.")
	fmt.Println()
}

func (r EntryReporter) printYearlyMDD(res eventstudy.Result) {
	fmt.Println("\n── 연도별 MDD ─────────────────────────────────────────────")
	fmt.Printf("%6s %10s %12s %12s %8s\n", "연도", "MDD(%)", "고점일", "저점일", "완전연도")
	for _, y := range res.YearlyMDD {
		mark := "○"
		if !y.Complete {
			mark = "부분"
		}
		fmt.Printf("%6d %10.2f %12s %12s %8s\n", y.Year, y.Pct,
			y.PeakDate.Format("2006-01-02"), y.TroughDate.Format("2006-01-02"), mark)
	}
}

func (r EntryReporter) printEntries(res eventstudy.Result, cur string) {
	fmt.Println("\n── 진입 사례 ──────────────────────────────────────────────")
	if len(res.Entries) == 0 {
		fmt.Println("조건을 만족한 날이 없다. 임계값을 완화하거나 기간을 늘려서 다시 볼 것.")
		return
	}

	header := fmt.Sprintf("%12s %14s %10s %10s %7s %12s %10s",
		"매수일", "매수가", "고점대비", "MA대비", "RSI", "이후 최저일", "최저(%)")
	fmt.Println(header)
	fmt.Println(strings.Repeat("-", utf8.RuneCountInString(header)))
	for _, e := range res.Entries {
		fmt.Printf("%12s %14s %9.1f%% %9.1f%% %7.1f %12s %9.1f%%\n",
			e.BuyDate.Format("2006-01-02"), formatMoney(cur, e.Buy),
			e.VsPeakPct, e.VsMAPct, e.RSI,
			e.LowDate.Format("2006-01-02"), e.LowPct)
	}
}

func (r EntryReporter) printTargets(res eventstudy.Result) {
	fmt.Println("\n── 목표 수익률 도달 기간 (달력일) ─────────────────────────")
	fmt.Printf("%8s %10s %12s %12s %8s %8s\n", "목표", "달성", "평균", "중위값", "최소", "최대")
	for _, t := range res.Targets {
		if t.Summary.N == 0 {
			fmt.Printf("%7.0f%% %10s %12s %12s %8s %8s\n",
				t.Target, fmt.Sprintf("%d/%d", t.Achieved, t.Total), "—", "—", "—", "—")
			continue
		}
		fmt.Printf("%7.0f%% %10s %11.0f일 %11.0f일 %7.0f일 %7.0f일\n",
			t.Target, fmt.Sprintf("%d/%d", t.Achieved, t.Total),
			t.Summary.Mean, t.Summary.Median, t.Summary.Min, t.Summary.Max)
	}
	r.printCensored(res)
}

// printCensored names the unfinished positions. Leaving them out of the table
// without a word turns "not yet" into "never", and the fastest observations are
// usually the newest ones.
func (r EntryReporter) printCensored(res eventstudy.Result) {
	var notes []string
	for _, e := range res.Entries {
		for _, o := range e.Outcomes {
			if !o.Achieved {
				notes = append(notes, fmt.Sprintf("%s +%.0f%% (%d일째)",
					e.BuyDate.Format("2006-01-02"), o.Target, o.Days))
			}
		}
	}
	if len(notes) == 0 {
		return
	}
	fmt.Printf("\n미달성 %d건은 통계에서 제외했다 — 아직 시간이 안 된 미완결 관측치이지 느린 사례가 아니다:\n  %s\n",
		len(notes), strings.Join(notes, ", "))
}

func (r EntryReporter) printValuation(val *valuation.Summary) {
	fmt.Println("\n── PER / PBR ──────────────────────────────────────────────")
	if val == nil || val.PER.N == 0 {
		fmt.Println("밸류에이션 CSV 가 없어 생략했다 (data/{ticker}_per_pbr.csv).")
		fmt.Println("수집 절차: .claude/skills/value-entry-backtest/references/data-sources.md")
		return
	}

	fmt.Printf("%s → %s · 거래일 %d일 · 실제 확보 %.1f년",
		val.From.Format("2006-01-02"), val.To.Format("2006-01-02"), val.PER.N, val.CoveredYears)
	if val.Short() {
		fmt.Printf("  ※ 요청 %d년보다 짧다 — 표 제목에 그대로 쓸 것", val.RequestedYrs)
	}
	fmt.Println()

	fmt.Printf("\n%6s %10s %10s %10s %10s %10s\n", "지표", "평균", "중위값", "최소", "최대", "현재")
	for _, m := range []valuation.Metric{val.PER, val.PBR} {
		fmt.Printf("%6s %10.2f %10.2f %10.2f %10.2f %10.2f\n", m.Name,
			m.Mean, m.Median, m.Min, m.Max, m.Last)
	}

	fmt.Printf("\n%6s %12s %12s %12s %12s\n", "연도", "PER 중위", "PER 평균", "PBR 중위", "PBR 평균")
	for _, y := range val.Yearly {
		fmt.Printf("%6d %12.2f %12.2f %12.2f %12.2f\n", y.Year,
			y.PER.Median, y.PER.Mean, y.PBR.Median, y.PBR.Mean)
	}
}
