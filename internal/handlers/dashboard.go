package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/puppe1990/amarra-cais/pkg/amarra/view"
	"github.com/puppe1990/amarra-cais/pkg/cais"
	"github.com/puppe1990/amarra-cais/pkg/cais/flash"
	"github.com/puppe1990/amarra-cais/pkg/cais/httpx"
	"github.com/puppe1990/amarra-cais/pkg/cais/i18n"
	"github.com/puppe1990/amarra-cais/pkg/cais/meta"
	"github.com/puppe1990/amarra-cais/pkg/cais/session"

	"github.com/gestao-bem/temposync/internal/jornada"
	"github.com/gestao-bem/temposync/internal/models"
	"github.com/gestao-bem/temposync/internal/store"
)

type DashboardHandler struct {
	views   *view.Renderer
	store   store.Store
	site    meta.Site
	catalog *i18n.Catalog
	cfg     cais.Config
}

func NewDashboardHandler(views *view.Renderer, s store.Store, site meta.Site, catalog *i18n.Catalog, cfg cais.Config) *DashboardHandler {
	return &DashboardHandler{views: views, store: s, site: site, catalog: catalog, cfg: cfg}
}

var ptWeekdays = []string{"Domingo", "Segunda-feira", "Terça-feira", "Quarta-feira", "Quinta-feira", "Sexta-feira", "Sábado"}
var ptWeekdaysShort = []string{"Dom", "Seg", "Ter", "Qua", "Qui", "Sex", "Sáb"}
var ptMonths = []string{"Janeiro", "Fevereiro", "Março", "Abril", "Maio", "Junho", "Julho", "Agosto", "Setembro", "Outubro", "Novembro", "Dezembro"}

func saoPauloNow() time.Time {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		return time.Now().UTC()
	}
	return time.Now().In(loc)
}

func dayStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func toJP(punches []models.Punch) []jornada.Punch {
	out := make([]jornada.Punch, 0, len(punches))
	for _, p := range punches {
		out = append(out, jornada.Punch{At: p.HappenedAt})
	}
	return out
}

func greeting(hour int) string {
	switch {
	case hour < 12:
		return "Bom dia"
	case hour < 18:
		return "Boa tarde"
	default:
		return "Boa noite"
	}
}

func initials(email string) string {
	name := strings.Split(email, "@")[0]
	name = strings.ReplaceAll(name, ".", " ")
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return "?"
	}
	out := strings.ToUpper(string([]rune(parts[0])[0]))
	if len(parts) > 1 {
		out += strings.ToUpper(string([]rune(parts[len(parts)-1])[0]))
	}
	return out
}

var punchLabels = []string{
	"1º Registro • Entrada Manhã",
	"2º Registro • Início Almoço",
	"3º Registro • Retorno Almoço",
	"4º Registro • Saída Tarde",
}

type timelineItem struct {
	Label   string
	Clock   string
	Detail  string
	Pending bool
}

func (h *DashboardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	uid, ok := session.UserID(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	user, err := h.store.FindUserByID(uid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	now := saoPauloNow()
	start := dayStart(now)
	punches, err := h.store.ListPunches(uid, start, start.Add(24*time.Hour))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jp := toJP(punches)
	worked := jornada.Worked(jp, now)
	brk := jornada.BreakTime(jp)
	forecast, hasForecast := jornada.ForecastExit(jp, jornada.Goal)
	forecastMinus10, forecastPlus30, forecastAlert := "--:--", "--:--", "--:--"
	if hasForecast {
		forecastMinus10 = jornada.FmtClock(forecast.Add(-10 * time.Minute))
		forecastPlus30 = jornada.FmtClock(forecast.Add(30 * time.Minute))
		forecastAlert = jornada.FmtClock(forecast.Add(-15 * time.Minute))
	}

	items := make([]timelineItem, 0, len(punches)+1)
	for i, p := range punches {
		label := punchLabels[i]
		if i >= len(punchLabels) {
			label = fmt.Sprintf("%dº Registro", i+1)
		}
		items = append(items, timelineItem{Label: label, Clock: jornada.FmtClockS(p.HappenedAt.In(now.Location())), Detail: "Confirmado • " + now.Location().String()})
	}
	if hasForecast && len(punches) < 4 {
		items = append(items, timelineItem{Label: fmt.Sprintf("%dº Registro • Saída Tarde (Projeção)", len(punches)+1), Clock: jornada.FmtClockS(forecast), Detail: "Previsão automatizada para bater 8h líquidas", Pending: true})
	}

	week, weekTotal := h.weekSummary(uid, now)
	monthBank := h.monthBank(uid, now)
	breakNote := "Intervalo padrão: 01h00"
	if brk > time.Hour {
		breakNote = fmt.Sprintf("Recomendado: 01h00 (+%dm na saída)", int((brk - time.Hour).Minutes()))
	}

	firstName := strings.Split(user.Email, "@")[0]
	writeView(w, r, h.views, h.cfg, "app", "dashboard", amarraData(r, h.site, map[string]any{
		"Title":           h.catalog.T("dashboard.title"),
		"ActiveNav":       "dashboard",
		"UserEmail":       user.Email,
		"UserName":        firstName,
		"UserInitials":    initials(user.Email),
		"Greeting":        greeting(now.Hour()),
		"TodayLong":       fmt.Sprintf("%s, %d de %s de %d", ptWeekdays[now.Weekday()], now.Day(), ptMonths[now.Month()-1], now.Year()),
		"NowClock":        now.Format("15:04:05"),
		"HasPunches":      len(punches) > 0,
		"PunchCount":      len(punches),
		"NextLabel":       jornada.NextLabel(len(punches)),
		"WorkedHM":        jornada.FmtHM(worked),
		"BreakHM":         jornada.FmtHM(brk),
		"BreakNote":       breakNote,
		"RemainingHM":     jornada.FmtHM(jornada.Remaining(worked, jornada.Goal)),
		"Percent":         jornada.Percent(worked, jornada.Goal),
		"HasForecast":     hasForecast,
		"ForecastClock":   jornada.FmtClock(forecast),
		"ForecastClockS":  jornada.FmtClockS(forecast),
		"ForecastMinus10": forecastMinus10,
		"ForecastPlus30":  forecastPlus30,
		"ForecastAlert":   forecastAlert,
		"Timeline":        items,
		"WeekDays":        week,
		"WeekTotal":       jornada.FmtSigned(weekTotal),
		"WeekTotalPos":    weekTotal >= 0,
		"WeekRange":       h.weekRange(now),
		"MonthBank":       jornada.FmtSigned(monthBank),
		"MonthBankPos":    monthBank >= 0,
		"TotalContacts":   0,
		"Env":             h.cfg.Env,
	}), 0)
}

type weekDay struct {
	Short string
	HM    string
	Pct   int
	Done  bool
	Today bool
}

func (h *DashboardHandler) weekSummary(uid int64, now time.Time) ([]weekDay, time.Duration) {
	monday := dayStart(now).AddDate(0, 0, -((int(now.Weekday()) + 6) % 7))
	out := make([]weekDay, 0, 5)
	var total time.Duration
	for i := 0; i < 5; i++ {
		day := monday.AddDate(0, 0, i)
		var worked time.Duration
		done := false
		if !day.After(dayStart(now)) {
			end := day.Add(24 * time.Hour)
			ref := now
			if day.Before(dayStart(now)) {
				ref = end
			}
			if punches, err := h.store.ListPunches(uid, day, end); err == nil && len(punches) > 0 {
				worked = jornada.Worked(toJP(punches), ref)
				done = true
				total += worked - jornada.Goal
			}
		}
		hm := "--:--"
		if done {
			hm = jornada.FmtHM(worked)
		}
		out = append(out, weekDay{Short: ptWeekdaysShort[day.Weekday()], HM: hm, Pct: jornada.Percent(worked, jornada.Goal), Done: done, Today: day.Equal(dayStart(now))})
	}
	return out, total
}

func (h *DashboardHandler) weekRange(now time.Time) string {
	monday := dayStart(now).AddDate(0, 0, -((int(now.Weekday()) + 6) % 7))
	friday := monday.AddDate(0, 0, 4)
	return fmt.Sprintf("%d a %d de %s", monday.Day(), friday.Day(), ptMonths[friday.Month()-1])
}

func (h *DashboardHandler) monthBank(uid int64, now time.Time) time.Duration {
	first := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	var total time.Duration
	for d := first; !d.After(dayStart(now)); d = d.AddDate(0, 0, 1) {
		if d.Weekday() == time.Saturday || d.Weekday() == time.Sunday {
			continue
		}
		punches, err := h.store.ListPunches(uid, d, d.Add(24*time.Hour))
		if err != nil || len(punches) == 0 {
			continue
		}
		ref := now
		if d.Before(dayStart(now)) {
			ref = d.Add(24 * time.Hour)
		}
		total += jornada.Worked(toJP(punches), ref) - jornada.Goal
	}
	return total
}

func (h *DashboardHandler) PunchPost(w http.ResponseWriter, r *http.Request) {
	uid, ok := session.UserID(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if err := httpx.ParseFormOrJSON(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	now := saoPauloNow()
	start := dayStart(now)
	existing, err := h.store.ListPunches(uid, start, start.Add(24*time.Hour))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := h.store.CreatePunch(uid, now, jornada.NextKind(len(existing))); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flash.Set(w, "notice", fmt.Sprintf("Ponto registrado às %s", now.Format("15:04:05")), h.cfg.CookieSecure())
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}
