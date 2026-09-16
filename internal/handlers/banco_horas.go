package handlers

import (
	"fmt"
	"net/http"
	"strconv"
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
	"github.com/gestao-bem/temposync/internal/store"
)

const (
	bancoTetoCiclo = 30 * time.Hour
	bancoTetoFolga = 8 * time.Hour
)

type BancoHorasHandler struct {
	views   *view.Renderer
	store   store.Store
	site    meta.Site
	catalog *i18n.Catalog
	cfg     cais.Config
}

func NewBancoHorasHandler(views *view.Renderer, s store.Store, site meta.Site, catalog *i18n.Catalog, cfg cais.Config) *BancoHorasHandler {
	return &BancoHorasHandler{views: views, store: s, site: site, catalog: catalog, cfg: cfg}
}

type bancoLancamento struct {
	Day        string
	Weekday    string
	Descricao  string
	Detalhe    string
	Kind       string
	KindLabel  string
	KindIcon   string
	KindCls    string
	DeltaHM    string
	DeltaNeg   bool
	SaldoAfter string
	SaldoPos   bool
	Status     string
	StatusIcon string
	StatusCls  string
	Today      bool
}

type bancoChartMonth struct {
	Label   string
	ExtraX  int
	CompX   int
	ExtraY  int
	CompY   int
	ExtraH  int
	CompH   int
	LineX   int
	LineY   int
	Current bool
}

type bancoSolicitacao struct {
	ID        string
	Kind      string
	KindIcon  string
	KindLabel string
	Day       string
	Hours     string
	Reason    string
	Status    string
	StatusCls string
	Pendente  bool
}

type bancoMonthAgg struct {
	label  string
	extra  time.Duration
	comp   time.Duration
	netRun time.Duration
}

func (h *BancoHorasHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	uid, ok := session.UserID(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	now := saoPauloNow()
	loc := now.Location()
	cyStart, cyEnd := semestre(now)
	days := h.collectDays(uid, cyStart, dayStart(now), now, loc)

	saldo := jornada.Saldo(days)
	creditos := jornada.Creditos(days)
	mesStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	mesDays := h.collectDays(uid, mesStart, dayStart(now), now, loc)
	mesSaldo := jornada.Saldo(mesDays)
	media := jornada.MediaDiaria(days)

	restantes := businessDaysLeft(now)
	proj := saldo + media*time.Duration(restantes)
	venc := cyEnd.AddDate(0, 1, 0)
	diasVenc := int(venc.Sub(dayStart(now)).Hours() / 24)

	yearDays := h.collectDays(uid, time.Date(now.Year(), 1, 1, 0, 0, 0, 0, loc), dayStart(now), now, loc)
	compensadoAno := jornada.Debitos(yearDays)
	folgasAno := 0
	for _, d := range yearDays {
		if d.Delta <= -4*time.Hour {
			folgasAno++
		}
	}

	tipo := r.URL.Query().Get("tipo")
	lancs := h.lancamentos(uid, days, tipo, now, loc)

	writeView(w, r, h.views, h.cfg, "app", "banco_horas", amarraData(r, h.site, map[string]any{
		"Title":           "Banco de Horas",
		"ActiveNav":       "banco-horas",
		"UserEmail":       h.userEmail(uid),
		"NowClock":        now.Format("15:04:05"),
		"TodayLong":       fmt.Sprintf("%s, %d de %s", ptWeekdays[now.Weekday()], now.Day(), ptMonths[now.Month()-1]),
		"CycleLabel":      fmt.Sprintf("%s – %s", fmt.Sprintf("%s/%d", ptMonths[cyStart.Month()-1][:3], cyStart.Year()), fmt.Sprintf("%s/%d", ptMonths[cyEnd.Month()-1][:3], cyEnd.Year())),
		"DueLabel":        fmt.Sprintf("%d de %s de %d", venc.Day(), ptMonths[venc.Month()-1], venc.Year()),
		"SaldoHM":         jornada.FmtSignedHM(saldo),
		"SaldoPos":        saldo >= 0,
		"SaldoMinutos":    int(saldo.Minutes()),
		"MesSaldoHM":      jornada.FmtSignedHM(mesSaldo),
		"TetoPct":         jornada.Pct(saldo, bancoTetoCiclo),
		"TetoLabel":       fmt.Sprintf("%dh", int(bancoTetoCiclo.Hours())),
		"ProjecaoHM":      jornada.FmtSignedHM(proj),
		"ProjecaoPct":     jornada.Pct(proj, bancoTetoCiclo),
		"ProjecaoLabel":   fmt.Sprintf("Cálculo estimado para %d/%s", lastBusinessDay(now).Day(), ptMonths[now.Month()-1][:3]),
		"VencDias":        diasVenc,
		"VencFolgas":      int(saldo / bancoTetoFolga),
		"VencPct":         jornada.Pct(vinteSeisSemanas(cyStart, cyEnd)-cyEnd.Sub(dayStart(now)), vinteSeisSemanas(cyStart, cyEnd)),
		"CompensadoAnoHM": jornada.FmtHM(compensadoAno),
		"FolgasAno":       folgasAno,
		"CreditosCicloHM": jornada.FmtHM(creditos),
		"Lancamentos":     lancs,
		"LancCount":       len(lancs),
		"LancTotal":       len(jornada.Lancamentos(days)),
		"Filtrando":       tipo,
		"ChartMonths":     chartMonths(days, cyStart, now),
		"Solicitacoes":    h.solicitacoes(uid),
		"Meses":           monthsOfCycle(cyStart),
		"MediaHM":         jornada.FmtSignedHM(media),
		"TodayISO":        now.Format("2006-01-02"),
		"NextMonthISO":    now.AddDate(0, 0, 7).Format("2006-01-02"),
	}), 0)
}

func semestre(now time.Time) (time.Time, time.Time) {
	if now.Month() <= time.June {
		return time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, now.Location()),
			time.Date(now.Year(), time.June, 30, 0, 0, 0, 0, now.Location())
	}
	return time.Date(now.Year(), time.July, 1, 0, 0, 0, 0, now.Location()),
		time.Date(now.Year(), time.December, 31, 0, 0, 0, 0, now.Location())
}

func vinteSeisSemanas(start, end time.Time) time.Duration {
	return end.AddDate(0, 1, 0).Sub(start)
}

func monthsOfCycle(start time.Time) []string {
	out := make([]string, 0, 6)
	for i := 0; i < 6; i++ {
		m := start.AddDate(0, i, 0)
		out = append(out, fmt.Sprintf("%s/%d", ptMonths[m.Month()-1][:3], m.Year()))
	}
	return out
}

func lastBusinessDay(now time.Time) time.Time {
	d := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, now.Location())
	for d.Weekday() == time.Saturday || d.Weekday() == time.Sunday {
		d = d.AddDate(0, 0, -1)
	}
	return d
}

func businessDaysLeft(now time.Time) int {
	n := 0
	for d := dayStart(now).AddDate(0, 0, 1); d.Month() == now.Month(); d = d.AddDate(0, 0, 1) {
		if isWorkday(d) {
			n++
		}
	}
	return n
}

func (h *BancoHorasHandler) collectDays(uid int64, from, to, now time.Time, loc *time.Location) []jornada.DayBalance {
	var out []jornada.DayBalance
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		if !isWorkday(d) {
			continue
		}
		if d.After(dayStart(now)) {
			continue
		}
		punches, err := h.store.ListPunches(uid, d, d.Add(24*time.Hour))
		if err != nil || len(punches) == 0 {
			continue
		}
		worked := jornada.Worked(toJP(punches, loc), refFor(d, now))
		out = append(out, jornada.DayBalance{Day: d, Delta: worked - jornada.Goal})
	}
	return out
}

func (h *BancoHorasHandler) lancamentos(uid int64, days []jornada.DayBalance, tipo string, now time.Time, loc *time.Location) []bancoLancamento {
	all := jornada.Lancamentos(days)
	out := make([]bancoLancamento, 0, len(all))
	today := dayStart(now)
	for _, l := range all {
		if tipo == "cred" && l.Kind != "credito" {
			continue
		}
		if tipo == "deb" && l.Kind != "debito" {
			continue
		}
		if tipo == "comp" && l.Kind != "compensacao" {
			continue
		}
		item := bancoLancamento{
			Day:        l.Day.Format("02/01/2006"),
			Weekday:    ptWeekdays[l.Day.Weekday()],
			DeltaHM:    jornada.FmtSignedHM(l.Delta),
			DeltaNeg:   l.Delta < 0,
			SaldoAfter: jornada.FmtSignedHM(l.SaldoAfter),
			SaldoPos:   l.SaldoAfter >= 0,
			Today:      l.Day.Equal(today),
		}
		switch l.Kind {
		case "credito":
			item.KindLabel, item.KindIcon, item.KindCls = "Crédito", "arrow_upward", "bg-secondary-container/40 text-on-secondary-container"
			item.Descricao = fmt.Sprintf("Jornada excedente de %s", jornada.FmtHM(l.Delta))
			item.Detalhe = "Crédito apurado pelas batidas do dia"
		case "compensacao":
			item.KindLabel, item.KindIcon, item.KindCls = "Compensação", "event_busy", "bg-primary-fixed/60 text-on-primary-fixed"
			item.Descricao = fmt.Sprintf("Folga / compensação de %s", jornada.FmtHM(l.Delta))
			item.Detalhe = "Débito integral de jornada usufruída"
		default:
			item.KindLabel, item.KindIcon, item.KindCls = "Débito", "arrow_downward", "bg-tertiary-fixed/60 text-on-tertiary-fixed"
			item.Descricao = fmt.Sprintf("Compensação de jornada reduzida (%s)", jornada.FmtHM(l.Delta))
			item.Detalhe = "Saída antecipada ou intervalo prolongado"
		}
		if item.Today {
			item.Status, item.StatusIcon, item.StatusCls = "Em andamento", "pending", "bg-tertiary-fixed/40 text-on-tertiary-fixed"
		} else {
			item.Status, item.StatusIcon, item.StatusCls = "Registro íntegro", "verified_user", "bg-surface-container text-on-surface"
		}
		out = append(out, item)
	}
	if len(out) > 10 {
		out = out[:10]
	}
	return out
}

func chartMonths(days []jornada.DayBalance, cyStart, now time.Time) []bancoChartMonth {
	aggs := make([]bancoMonthAgg, 6)
	for i := range aggs {
		m := cyStart.AddDate(0, i, 0)
		aggs[i].label = ptMonths[m.Month()-1][:3]
	}
	var running time.Duration
	for _, d := range days {
		idx := int(d.Day.Month()) - int(cyStart.Month())
		if d.Day.Year() != cyStart.Year() {
			idx += 12
		}
		if idx < 0 || idx >= len(aggs) {
			continue
		}
		e := jornada.EffectiveDelta(d.Delta)
		if e > 0 {
			aggs[idx].extra += e
		} else if e < 0 {
			aggs[idx].comp -= e
		}
		running += e
		aggs[idx].netRun = running
	}
	var maxV time.Duration
	for _, a := range aggs {
		if a.extra > maxV {
			maxV = a.extra
		}
		if a.comp > maxV {
			maxV = a.comp
		}
	}
	const (
		baseY = 170
		maxH  = 120
	)
	height := func(v time.Duration) int {
		if maxV <= 0 {
			return 0
		}
		h := int(v * maxH / maxV)
		if v > 0 && h < 4 {
			h = 4
		}
		return h
	}
	netLine := func(n time.Duration) int {
		if maxV <= 0 {
			return baseY
		}
		h := int(n * maxH / maxV)
		y := baseY - h
		if y < 10 {
			y = 10
		}
		if y > baseY {
			y = baseY
		}
		return y
	}
	out := make([]bancoChartMonth, 0, 6)
	for i, a := range aggs {
		x := 50 + i*115
		item := bancoChartMonth{
			Label:   a.label,
			ExtraX:  x,
			CompX:   x + 26,
			ExtraH:  height(a.extra),
			CompH:   height(a.comp),
			LineX:   x + 24,
			LineY:   netLine(a.netRun),
			Current: a.label == ptMonths[now.Month()-1][:3] && i == int(now.Month())-int(cyStart.Month()),
		}
		item.ExtraY = baseY - item.ExtraH
		item.CompY = baseY - item.CompH
		out = append(out, item)
	}
	return out
}

func (h *BancoHorasHandler) solicitacoes(uid int64) []bancoSolicitacao {
	list, err := h.store.ListRequests(uid)
	if err != nil {
		return nil
	}
	out := make([]bancoSolicitacao, 0, len(list))
	for _, r := range list {
		item := bancoSolicitacao{
			ID:     fmt.Sprintf("%d", r.ID),
			Day:    r.Day.In(saoPauloNow().Location()).Format("02/01/2006"),
			Hours:  jornada.FmtHM(time.Duration(r.Minutes) * time.Minute),
			Reason: r.Reason,
		}
		switch r.Kind {
		case "ajuste":
			item.Kind, item.KindLabel, item.KindIcon = "ajuste", "Ajuste de ponto", "edit_calendar"
		default:
			item.Kind, item.KindLabel, item.KindIcon = "folga", "Folga compensatória", "event_available"
		}
		switch r.Status {
		case "aprovada":
			item.Status, item.StatusCls = "Aprovada", "bg-secondary-container/50 text-on-secondary-container"
		case "cancelada":
			item.Status, item.StatusCls = "Cancelada", "bg-surface-container text-on-surface-variant"
		default:
			item.Status, item.StatusCls, item.Pendente = "Pendente", "bg-tertiary-fixed/60 text-on-tertiary-fixed", true
		}
		out = append(out, item)
	}
	if len(out) > 8 {
		out = out[:8]
	}
	return out
}

// SolicitacoesPost cria/decide solicitações de folga (auto-gestão).
func (h *BancoHorasHandler) SolicitacoesPost(w http.ResponseWriter, r *http.Request) {
	uid, ok := session.UserID(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if err := httpx.ParseFormOrJSON(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	action := r.FormValue("acao")
	switch action {
	case "criar":
		day, err := time.ParseInLocation("2006-01-02", r.FormValue("day"), saoPauloNow().Location())
		if err != nil {
			flash.Set(w, "error", "Data inválida para a solicitação", h.cfg.CookieSecure())
			http.Redirect(w, r, "/banco-horas", http.StatusSeeOther)
			return
		}
		hours, _ := strconv.Atoi(r.FormValue("hours"))
		if hours < 1 || hours > 80 {
			flash.Set(w, "error", "Informe entre 1 e 80 horas", h.cfg.CookieSecure())
			http.Redirect(w, r, "/banco-horas", http.StatusSeeOther)
			return
		}
		kind := "folga"
		if r.FormValue("kind") == "ajuste" {
			kind = "ajuste"
		}
		reason := strings.TrimSpace(r.FormValue("reason"))
		if _, err := h.store.CreateRequest(uid, kind, day, hours*60, reason); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		flash.Set(w, "notice", "Solicitação registrada", h.cfg.CookieSecure())
	case "aprovar", "cancelar":
		id, err := strconv.ParseInt(r.FormValue("id"), 10, 64)
		if err != nil {
			http.Error(w, "id inválido", http.StatusUnprocessableEntity)
			return
		}
		status := "aprovada"
		if action == "cancelar" {
			status = "cancelada"
		}
		if err := h.store.DecideRequest(uid, id, status); err != nil {
			flash.Set(w, "error", "Solicitação não encontrada ou já decidida", h.cfg.CookieSecure())
			http.Redirect(w, r, "/banco-horas", http.StatusSeeOther)
			return
		}
		if status == "aprovada" {
			h.registrarFolga(uid, id, w, r)
			return
		}
		flash.Set(w, "notice", "Solicitação cancelada", h.cfg.CookieSecure())
	default:
		http.Error(w, "ação inválida", http.StatusUnprocessableEntity)
		return
	}
	http.Redirect(w, r, "/banco-horas", http.StatusSeeOther)
}

// registrarFolga materializa a folga aprovada como batida kind=folga.
func (h *BancoHorasHandler) registrarFolga(uid, requestID int64, w http.ResponseWriter, r *http.Request) {
	list, err := h.store.ListRequests(uid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, req := range list {
		if req.ID != requestID {
			continue
		}
		at := req.Day.In(saoPauloNow().Location())
		if _, err := h.store.CreateFolga(uid, at, req.Minutes, req.Reason); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		flash.Set(w, "notice", "Folga aprovada e lançada no banco de horas", h.cfg.CookieSecure())
		http.Redirect(w, r, "/banco-horas", http.StatusSeeOther)
		return
	}
	flash.Set(w, "error", "Solicitação não encontrada", h.cfg.CookieSecure())
	http.Redirect(w, r, "/banco-horas", http.StatusSeeOther)
}

func (h *BancoHorasHandler) userEmail(uid int64) string {
	u, err := h.store.FindUserByID(uid)
	if err != nil {
		return ""
	}
	return u.Email
}
