package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/puppe1990/amarra-cais/pkg/amarra/view"
	"github.com/puppe1990/amarra-cais/pkg/cais"
	"github.com/puppe1990/amarra-cais/pkg/cais/i18n"
	"github.com/puppe1990/amarra-cais/pkg/cais/meta"
	"github.com/puppe1990/amarra-cais/pkg/cais/session"

	"github.com/gestao-bem/temposync/internal/jornada"
	"github.com/gestao-bem/temposync/internal/store"
)

type MetricasHandler struct {
	views   *view.Renderer
	store   store.Store
	site    meta.Site
	catalog *i18n.Catalog
	cfg     cais.Config
}

func NewMetricasHandler(views *view.Renderer, s store.Store, site meta.Site, catalog *i18n.Catalog, cfg cais.Config) *MetricasHandler {
	return &MetricasHandler{views: views, store: s, site: site, catalog: catalog, cfg: cfg}
}

type monthStats struct {
	month       time.Time
	business    int
	worked      time.Duration
	breaks      time.Duration
	extras      time.Duration
	bank        time.Duration
	withPunches int
	onTime      int
	lates       int
	faults      int
	longDays    int
	doneDays    int
}

type distBar struct {
	Day   string
	Pct   int
	Class string
	Tip   string
	Bold  bool
}

type donutSeg struct {
	Len      string
	Off      string
	Class    string
	Label    string
	Value    string
	PctShort string
	Dot      string
}

type weekdayPattern struct {
	Day   string
	Entry string
	Exit  string
	Load  string
	Left  string
	Width string
	Today bool
}

type insightItem struct {
	Icon  string
	Bg    string
	Title string
	Text  string
}

type historyRow struct {
	Month     string
	Days      string
	Expected  string
	Done      string
	Balance   string
	BalanceCl string
	ExtraRate string
	Abonos    string
	Status    string
	StatusCls string
}

func (h *MetricasHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	uid, ok := session.UserID(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	now := saoPauloNow()
	loc := now.Location()
	month := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	st := h.collect(uid, month, dayStart(now).Add(24*time.Hour), now)

	avg := time.Duration(0)
	if st.withPunches > 0 {
		avg = st.worked / time.Duration(st.withPunches)
	}
	punct := 100.0
	if st.withPunches > 0 {
		punct = float64(st.onTime) * 100 / float64(st.withPunches)
	}

	writeView(w, r, h.views, h.cfg, "app", "metricas", amarraData(r, h.site, map[string]any{
		"Title":       "Análise & Métricas",
		"ActiveNav":   "metricas",
		"UserEmail":   h.userEmail(uid),
		"NowClock":    now.Format("15:04:05"),
		"TodayLong":   fmt.Sprintf("%s, %d de %s", ptWeekdays[now.Weekday()], now.Day(), ptMonths[now.Month()-1]),
		"MonthLabel":  fmt.Sprintf("%s %d", ptMonths[month.Month()-1], month.Year()),
		"DailyAvgHM":  jornada.FmtHM(avg),
		"Punctuality": fmt.Sprintf("%.1f", punct),
		"BankMonth":   jornada.FmtSigned(st.bank),
		"BankPos":     st.bank >= 0,
		"Score":       score(st),
		"DistBars":    h.distBars(uid, month, dayStart(now).Add(24*time.Hour), now),
		"Donut":       donut(st),
		"Pattern":     h.pattern(uid, month, dayStart(now), loc),
		"Insights":    insights(st),
		"History":     h.history(uid, now, 6),
	}), 0)
}

func (h *MetricasHandler) collect(uid int64, month, until, now time.Time) monthStats {
	st := monthStats{month: month}
	loc := month.Location()
	for d := month; d.Before(until); d = d.AddDate(0, 0, 1) {
		if !isWorkday(d) {
			continue
		}
		st.business++
		punches, err := h.store.ListPunches(uid, d, d.Add(24*time.Hour))
		if err != nil || len(punches) == 0 {
			if d.Before(dayStart(now)) {
				st.faults++
			}
			continue
		}
		st.withPunches++
		st.doneDays++
		worked := jornada.Worked(toJP(punches, loc), refFor(d, now))
		st.worked += worked
		st.breaks += jornada.BreakTime(toJP(punches, loc))
		if first := punches[0].HappenedAt.In(loc); first.Hour()*60+first.Minute() <= 8*60+10 {
			st.onTime++
		} else {
			st.lates++
		}
		if extra := worked - jornada.Goal; extra > 0 {
			st.extras += extra
			if extra > time.Hour {
				st.longDays++
			}
		}
		st.bank += worked - jornada.Goal
	}
	return st
}

func score(st monthStats) int {
	s := 100 - st.faults*10 - st.longDays*5 - st.lates*2
	if s < 0 {
		return 0
	}
	return s
}

func donut(st monthStats) []donutSeg {
	circ := 238.76
	segs := []struct {
		d     time.Duration
		class string
		label string
		dot   string
	}{
		{st.worked - st.breaks, "text-primary", "Expediente Focado", "bg-primary"},
		{st.breaks, "text-secondary", "Intervalos & Almoço", "bg-secondary"},
		{st.extras, "text-tertiary-fixed-dim", "Horas Extras", "bg-tertiary-fixed-dim"},
	}
	sum := time.Duration(0)
	for _, s := range segs {
		if s.d > 0 {
			sum += s.d
		}
	}
	if sum <= 0 {
		return []donutSeg{
			{Len: "0 238.76", Off: "0", Class: "text-surface-container-high", Label: "Sem dados", Value: "0%", PctShort: "0%", Dot: "bg-surface-container-high"},
		}
	}
	out := make([]donutSeg, 0, len(segs))
	start0 := 0.0
	for _, s := range segs {
		pct := float64(s.d) / float64(sum) * 100
		if pct < 0 {
			pct = 0
		}
		length := pct * circ / 100
		out = append(out, donutSeg{
			Len:      fmt.Sprintf("%.1f %.2f", length, circ),
			Off:      fmt.Sprintf("%.1f", -start0),
			Class:    s.class,
			Label:    s.label,
			Value:    fmt.Sprintf("%s (%s)", fmt.Sprintf("%.0f%%", pct), jornada.FmtHM(s.d)),
			PctShort: fmt.Sprintf("%.0f%%", pct),
			Dot:      s.dot,
		})
		start0 += length
	}
	return out
}

func insights(st monthStats) []insightItem {
	out := []insightItem{}
	if st.breaks > 0 && st.withPunches > 0 {
		avgBreak := st.breaks / time.Duration(st.withPunches)
		out = append(out, insightItem{Icon: "restaurant", Bg: "bg-primary-fixed text-primary", Title: "Intervalos de Almoço",
			Text: fmt.Sprintf("Cumprimento médio de %s, garantindo recuperação adequada e conformidade CLT Art. 71.", jornada.FmtHM(avgBreak))})
	}
	switch {
	case st.bank >= 0:
		out = append(out, insightItem{Icon: "energy_savings_leaf", Bg: "bg-secondary-container text-secondary", Title: "Banco Saudável",
			Text: fmt.Sprintf("Saldo de %s neste mês. Programe folga para aproveitar o crédito antes do ciclo semestral.", jornada.FmtSigned(st.bank))})
	default:
		out = append(out, insightItem{Icon: "event_busy", Bg: "bg-tertiary-fixed text-tertiary", Title: "Risco de Débito",
			Text: fmt.Sprintf("Saldo de %s neste mês. Compense nos próximos dias para manter o banco em dia com o CCT.", jornada.FmtSigned(st.bank))})
	}
	switch {
	case st.longDays > 1:
		out = append(out, insightItem{Icon: "local_fire_department", Bg: "bg-tertiary-fixed text-tertiary", Title: "Acúmulo de Extras",
			Text: fmt.Sprintf("%d dias com mais de 1h extra. Distribua melhor a carga para preservar recuperação.", st.longDays)})
	case st.faults > 0:
		out = append(out, insightItem{Icon: "event_busy", Bg: "bg-error-container text-on-error-container", Title: "Registros Pendentes",
			Text: fmt.Sprintf("%d dia(s) sem registro neste mês. Regularize via ajuste para evitar descontos.", st.faults)})
	default:
		out = append(out, insightItem{Icon: "bolt", Bg: "bg-secondary-container text-secondary", Title: "Ritmo Constante",
			Text: "Sem atrasos ou faltas relevantes: sua consistência mantém o score de equilíbrio alto."})
	}
	return out
}

func (h *MetricasHandler) distBars(uid int64, month, until, now time.Time) []distBar {
	loc := month.Location()
	out := []distBar{}
	for d := month; d.Before(until); d = d.AddDate(0, 0, 1) {
		if !isWorkday(d) {
			continue
		}
		bar := distBar{Day: fmt.Sprintf("%02d", d.Day()), Pct: 1, Class: "bg-surface-container-high", Tip: fmt.Sprintf("%02d/%s: sem registro", d.Day(), ptMonths[d.Month()-1][:3])}
		if punches, err := h.store.ListPunches(uid, d, d.Add(24*time.Hour)); err == nil && len(punches) > 0 {
			worked := jornada.Worked(toJP(punches, loc), refFor(d, now))
			bal := worked - jornada.Goal
			bar.Pct = int(worked * 100 / (10 * time.Hour))
			if bar.Pct < 6 {
				bar.Pct = 6
			}
			if bar.Pct > 96 {
				bar.Pct = 96
			}
			bar.Tip = fmt.Sprintf("%02d/%s: %s", d.Day(), ptMonths[d.Month()-1][:3], jornada.FmtHM(worked))
			switch {
			case bal > 10*time.Minute:
				bar.Class = "bg-secondary"
			case bal < -10*time.Minute:
				bar.Class = "bg-tertiary-container"
			default:
				bar.Class = "bg-primary-container"
			}
		}
		if d.Equal(dayStart(now)) {
			bar.Class = "bg-primary animate-pulse"
			bar.Bold = true
		}
		out = append(out, bar)
	}
	return out
}

func (h *MetricasHandler) pattern(uid int64, month, today time.Time, loc *time.Location) []weekdayPattern {
	type acc struct {
		entryMins, exitMins, loadMins int
		n                             int
	}
	buckets := make([]acc, 7)
	for d := month; d.Before(today.AddDate(0, 0, 1)); d = d.AddDate(0, 0, 1) {
		if !isWorkday(d) {
			continue
		}
		punches, err := h.store.ListPunches(uid, d, d.Add(24*time.Hour))
		if err != nil || len(punches) == 0 {
			continue
		}
		b := &buckets[d.Weekday()]
		first := punches[0].HappenedAt.In(loc)
		b.entryMins += first.Hour()*60 + first.Minute()
		last := punches[len(punches)-1].HappenedAt.In(loc)
		b.exitMins += last.Hour()*60 + last.Minute()
		b.loadMins += int(jornada.Worked(toJP(punches, loc), refFor(d, today)).Minutes())
		b.n++
	}
	out := make([]weekdayPattern, 0, 5)
	for wd := time.Monday; wd <= time.Friday; wd++ {
		b := buckets[wd]
		p := weekdayPattern{Day: ptWeekdays[wd], Entry: "--:--", Exit: "--:--", Load: "--:--", Left: "0", Width: "0"}
		if b.n > 0 {
			em, xm, lm := b.entryMins/b.n, b.exitMins/b.n, b.loadMins/b.n
			p.Entry = fmt.Sprintf("%02d:%02d", em/60, em%60)
			p.Exit = fmt.Sprintf("%02d:%02d", xm/60, xm%60)
			p.Load = jornada.FmtHM(time.Duration(lm) * time.Minute)
			left := float64(em-360) / 840 * 100
			width := float64(xm-em) / 840 * 100
			if left < 0 {
				left = 0
			}
			if width < 5 {
				width = 5
			}
			if left+width > 100 {
				width = 100 - left
			}
			p.Left = fmt.Sprintf("%.0f", left)
			p.Width = fmt.Sprintf("%.0f", width)
		}
		p.Today = today.Weekday() == wd
		out = append(out, p)
	}
	return out
}

func (h *MetricasHandler) history(uid int64, now time.Time, months int) []historyRow {
	out := make([]historyRow, 0, months)
	first := dayStart(now).AddDate(0, 0, 1)
	for i := 0; i < months; i++ {
		month := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, -i, 0)
		until := month.AddDate(0, 1, 0)
		partial := until.After(first)
		if partial {
			until = first
		}
		st := h.collect(uid, month, until, now)
		expected := time.Duration(st.business) * jornada.Goal
		row := historyRow{
			Month:     fmt.Sprintf("%s / %d", ptMonths[month.Month()-1], month.Year()),
			Days:      fmt.Sprintf("%d", st.business),
			Expected:  jornada.FmtHM(expected),
			Done:      jornada.FmtHM(st.worked),
			Balance:   jornada.FmtSigned(st.bank),
			Abonos:    "0h (0 dias)",
			Status:    "Fechado",
			StatusCls: "bg-secondary-container text-on-secondary-container",
		}
		if expected > 0 {
			row.ExtraRate = fmt.Sprintf("%.1f%%", float64(st.extras)/float64(expected)*100)
		} else {
			row.ExtraRate = "0.0%"
		}
		switch {
		case st.bank > 10*time.Minute:
			row.BalanceCl = "text-secondary"
		case st.bank < -10*time.Minute:
			row.BalanceCl = "text-error"
		default:
			row.BalanceCl = "text-on-surface"
		}
		if partial {
			row.Status = "Em aberto"
			row.StatusCls = "bg-primary-fixed text-on-primary-fixed"
		} else if st.doneDays == 0 {
			row.Status = "Sem registros"
			row.StatusCls = "bg-surface-container text-on-surface-variant"
		}
		out = append(out, row)
	}
	return out
}

func (h *MetricasHandler) userEmail(uid int64) string {
	u, err := h.store.FindUserByID(uid)
	if err != nil {
		return ""
	}
	return u.Email
}
