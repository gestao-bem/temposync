package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/puppe1990/amarra-cais/pkg/amarra/view"
	"github.com/puppe1990/amarra-cais/pkg/cais"
	"github.com/puppe1990/amarra-cais/pkg/cais/i18n"
	"github.com/puppe1990/amarra-cais/pkg/cais/meta"
	"github.com/puppe1990/amarra-cais/pkg/cais/session"

	"github.com/gestao-bem/temposync/internal/jornada"
	"github.com/gestao-bem/temposync/internal/models"
	"github.com/gestao-bem/temposync/internal/store"
)

type EspelhoHandler struct {
	views   *view.Renderer
	store   store.Store
	site    meta.Site
	catalog *i18n.Catalog
	cfg     cais.Config
}

func NewEspelhoHandler(views *view.Renderer, s store.Store, site meta.Site, catalog *i18n.Catalog, cfg cais.Config) *EspelhoHandler {
	return &EspelhoHandler{views: views, store: s, site: site, catalog: catalog, cfg: cfg}
}

type dayRow struct {
	Top        string
	Sub        string
	SubClass   string
	Today      bool
	In1        string
	BrkOut     string
	BrkBack    string
	Out2       string
	Out2State  string
	Liquid     string
	Balance    string
	BalancePos bool
	BalanceNeg bool
	StatusIcon string
	StatusText string
	StatusCls  string
}

type chartBar struct {
	X     int
	Y     int
	H     int
	Class string
	Day   string
	Bold  bool
}

func isWorkday(d time.Time) bool {
	return d.Weekday() != time.Saturday && d.Weekday() != time.Sunday
}

func (h *EspelhoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	uid, ok := session.UserID(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	now := saoPauloNow()
	loc := now.Location()
	month := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	if q := r.URL.Query().Get("month"); q != "" {
		if t, err := time.ParseInLocation("2006-01", q, loc); err == nil {
			month = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, loc)
		}
	}
	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	monthEnd := month.AddDate(0, 1, 0)
	lastDay := monthEnd
	if month.Year() == now.Year() && month.Month() == now.Month() {
		lastDay = dayStart(now).Add(24 * time.Hour)
	}

	var rows []dayRow
	var done, bank time.Duration
	var onTime, withPunches, faults int
	businessDays := 0
	for d := month; d.Before(monthEnd); d = d.AddDate(0, 0, 1) {
		if !isWorkday(d) {
			continue
		}
		businessDays++
		if !d.Before(lastDay) {
			continue
		}
		punches, err := h.store.ListPunches(uid, d, d.Add(24*time.Hour))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		row := h.buildRow(punches, d, now, loc)
		if len(punches) > 0 {
			withPunches++
			worked := jornada.Worked(toJP(punches, loc), refFor(d, now))
			done += worked
			bank += worked - jornada.Goal
			if first := punches[0].HappenedAt.In(loc); first.Hour()*60+first.Minute() <= 8*60+10 {
				onTime++
			}
		} else if d.Before(dayStart(now)) {
			faults++
		}
		if q == "" || strings.Contains(strings.ToLower(row.Top+" "+row.Sub+" "+row.StatusText+" "+row.In1+row.Out2), q) {
			rows = append([]dayRow{row}, rows...)
		}
	}

	punct := 100.0
	if withPunches > 0 {
		punct = float64(onTime) * 100 / float64(withPunches)
	}
	prev := month.AddDate(0, -1, 0).Format("2006-01")
	next := ""
	if month.Before(time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)) {
		next = month.AddDate(0, 1, 0).Format("2006-01")
	}

	writeView(w, r, h.views, h.cfg, "app", "espelho", amarraData(r, h.site, map[string]any{
		"Title":        "Espelho de Ponto",
		"ActiveNav":    "espelho",
		"UserEmail":    h.userEmail(uid),
		"NowClock":     now.Format("15:04:05"),
		"TodayLong":    fmt.Sprintf("%s, %d de %s", ptWeekdays[now.Weekday()], now.Day(), ptMonths[now.Month()-1]),
		"MonthLabel":   fmt.Sprintf("%s %d", ptMonths[month.Month()-1], month.Year()),
		"MonthParam":   month.Format("2006-01"),
		"PrevMonth":    prev,
		"NextMonth":    next,
		"HasNext":      next != "",
		"Query":        r.URL.Query().Get("q"),
		"Days":         rows,
		"DayCount":     len(rows),
		"ExpectedHM":   jornada.FmtHM(time.Duration(businessDays) * jornada.Goal),
		"BusinessDays": businessDays,
		"DoneHM":       hmSplit(done),
		"DoneExtra":    jornada.FmtSigned(bank),
		"BankMonth":    jornada.FmtSigned(bank),
		"BankPos":      bank >= 0,
		"Punctuality":  fmt.Sprintf("%.1f", punct),
		"Faults":       faults,
		"ChartBars":    h.chartBars(uid, month, lastDay, loc),
		"MonthShort":   ptMonths[month.Month()-1][:3],
	}), 0)
}

func refFor(day, now time.Time) time.Time {
	if day.Before(dayStart(now)) {
		return day.Add(24 * time.Hour)
	}
	return now
}

func hmSplit(d time.Duration) map[string]string {
	if d < 0 {
		d = -d
	}
	return map[string]string{
		"H": fmt.Sprintf("%02d", int(d.Hours())),
		"M": fmt.Sprintf("%02d", int(d.Minutes())%60),
	}
}

func (h *EspelhoHandler) userEmail(uid int64) string {
	u, err := h.store.FindUserByID(uid)
	if err != nil {
		return ""
	}
	return u.Email
}

func (h *EspelhoHandler) buildRow(punches []models.Punch, day, now time.Time, loc *time.Location) dayRow {
	top := fmt.Sprintf("%02d/%02d %s", day.Day(), day.Month(), ptWeekdaysShort[day.Weekday()])
	today := day.Equal(dayStart(now))
	empty := dayRow{Top: top, In1: "--:--", BrkOut: "--:--", BrkBack: "--:--", Out2: "--:--", Liquid: "--:--", Balance: "--:--"}
	if len(punches) == 0 {
		empty.Sub = "Sem registro"
		empty.SubClass = "text-error"
		empty.StatusIcon = "event_busy"
		empty.StatusText = "Sem registro"
		empty.StatusCls = "bg-error-container/60 text-on-error-container"
		if today {
			empty.Sub = "Aguardando batidas"
			empty.SubClass = ""
			empty.StatusIcon = "schedule"
			empty.StatusText = "Aguardando batidas"
			empty.StatusCls = "bg-surface-container text-on-surface-variant"
		}
		empty.Today = today
		return empty
	}
	clk := func(t time.Time) string { return t.In(loc).Format("15:04") }
	row := dayRow{Top: top, Sub: "Expediente", In1: clk(punches[0].HappenedAt)}
	if len(punches) > 1 {
		row.BrkOut = clk(punches[1].HappenedAt)
	}
	if len(punches) > 2 {
		row.BrkBack = clk(punches[2].HappenedAt)
	}
	worked := jornada.Worked(toJP(punches, loc), refFor(day, now))
	row.Liquid = jornada.FmtHM(worked)
	bal := worked - jornada.Goal
	row.Balance = jornada.FmtSigned(bal)
	row.BalancePos = bal > 10*time.Minute
	row.BalanceNeg = bal < -10*time.Minute
	switch {
	case today && len(punches) < 4:
		row.Sub = "Hoje • Em Andamento"
		row.Out2 = "Em curso"
		row.Out2State = "open"
		if fc, ok := jornada.ForecastExit(toJP(punches, loc), jornada.Goal); ok {
			row.StatusIcon = "timelapse"
			row.StatusText = "Previsto Saída: " + fc.In(loc).Format("15:04")
			row.StatusCls = "bg-primary/10 text-primary"
		}
	case row.BalancePos:
		row.StatusIcon = "more_time"
		row.StatusText = "Hora Extra Computada"
		row.StatusCls = "bg-secondary-container/40 text-on-secondary-container"
	case row.BalanceNeg:
		row.StatusIcon = "hourglass_bottom"
		row.StatusText = "Débito em Banco"
		row.StatusCls = "bg-error-container/60 text-on-error-container"
	default:
		row.StatusIcon = "task_alt"
		row.StatusText = "Jornada Regular"
		row.StatusCls = "bg-surface-container text-on-surface-variant"
	}
	if len(punches) > 3 {
		row.Out2 = clk(punches[3].HappenedAt)
	}
	row.Today = today
	return row
}

func (h *EspelhoHandler) chartBars(uid int64, month, lastDay time.Time, loc *time.Location) []chartBar {
	var days []time.Time
	for d := month; d.Before(lastDay); d = d.AddDate(0, 0, 1) {
		if isWorkday(d) {
			days = append(days, d)
		}
	}
	out := make([]chartBar, 0, len(days))
	for i, d := range days {
		x := 50
		if len(days) > 1 {
			x = 50 + i*410/(len(days)-1)
		}
		bar := chartBar{X: x, Y: 65, H: 2, Class: "fill-outline-variant", Day: fmt.Sprintf("%02d", d.Day())}
		punches, err := h.store.ListPunches(uid, d, d.Add(24*time.Hour))
		if err == nil && len(punches) > 0 {
			bal := jornada.Worked(toJP(punches, loc), refFor(d, saoPauloNow())) - jornada.Goal
			mins := int(bal.Minutes())
			hh := mins * 47 / 90
			if hh < 0 {
				hh = -hh
			}
			if hh < 4 {
				hh = 4
			}
			if hh > 47 {
				hh = 47
			}
			bar.H = hh
			bar.Class = "fill-secondary"
			if mins < -10 {
				bar.Class = "fill-error"
			} else if mins <= 10 {
				bar.Class = "fill-outline-variant"
			}
			if mins >= 0 {
				bar.Y = 65 - hh
			}
		}
		if d.Equal(dayStart(saoPauloNow())) {
			bar.Class = "fill-primary animate-pulse"
			bar.Bold = true
		}
		out = append(out, bar)
	}
	return out
}
