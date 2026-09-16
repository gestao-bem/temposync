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
	"github.com/gestao-bem/temposync/internal/store"
)

type ConfiguracoesHandler struct {
	views   *view.Renderer
	store   store.Store
	site    meta.Site
	catalog *i18n.Catalog
	cfg     cais.Config
}

func NewConfiguracoesHandler(views *view.Renderer, s store.Store, site meta.Site, catalog *i18n.Catalog, cfg cais.Config) *ConfiguracoesHandler {
	return &ConfiguracoesHandler{views: views, store: s, site: site, catalog: catalog, cfg: cfg}
}

type gridDay struct {
	Key       string
	Name      string
	Status    string
	StatusCls string
	In        string
	Break     string
	Out       string
	Liquid    string
	Off       bool
}

type configDefaults struct {
	contract      string
	compSat       bool
	lunch         string
	lunchFlex     bool
	breakExtra    bool
	alertMaster   bool
	alertLead     string
	chanSound     bool
	chanPush      bool
	chanTelegram  bool
	alertToleranc bool
	cycle         string
}

func defaultConfig() configDefaults {
	return configDefaults{
		contract:      "clt-44",
		compSat:       true,
		lunch:         "60",
		lunchFlex:     true,
		breakExtra:    true,
		alertMaster:   true,
		alertLead:     "15",
		chanSound:     true,
		chanPush:      true,
		chanTelegram:  false,
		alertToleranc: true,
		cycle:         "semestral",
	}
}

var configGridDefaults = map[string][3]string{
	"mon": {"08:00", "01:00", "17:48"},
	"tue": {"08:00", "01:00", "17:48"},
	"wed": {"08:00", "01:00", "17:48"},
	"thu": {"08:00", "01:00", "17:48"},
	"fri": {"08:00", "01:00", "17:00"},
}

var configGridNames = map[string]string{
	"mon": "Segunda-feira",
	"tue": "Terça-feira",
	"wed": "Quarta-feira",
	"thu": "Quinta-feira",
	"fri": "Sexta-feira",
}

var (
	validContracts = map[string]bool{"clt-44": true, "clt-40": true, "pj": true}
	validLunch     = map[string]bool{"60": true, "75": true, "90": true}
	validLead      = map[string]bool{"5": true, "10": true, "15": true, "30": true}
	validCycles    = map[string]bool{"mensal": true, "semestral": true, "anual": true}
)

func boolSetting(v string, def bool, ok bool) bool {
	if !ok {
		return def
	}
	return v == "on" || v == "true" || v == "1"
}

func pickSetting(v string, valid map[string]bool, def string) string {
	if valid[v] {
		return v
	}
	return def
}

func (h *ConfiguracoesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	uid, ok := session.UserID(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	settings, err := h.store.GetSettings(uid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	def := defaultConfig()
	get := func(key, dv string) string {
		if v, ok := settings[key]; ok {
			return v
		}
		return dv
	}
	getBool := func(key string, dv bool) bool {
		v, ok := settings[key]
		return boolSetting(v, dv, ok)
	}

	grid := make([]gridDay, 0, 5)
	var weekTotal time.Duration
	for _, key := range []string{"mon", "tue", "wed", "thu", "fri"} {
		d := configGridDefaults[key]
		inS := get("grid_"+key+"_in", d[0])
		brkS := get("grid_"+key+"_break", d[1])
		outS := get("grid_"+key+"_out", d[2])
		in, okIn := jornada.ParseHHMM(inS)
		brk, okBrk := jornada.ParseHHMM(brkS)
		out, okOut := jornada.ParseHHMM(outS)
		liquid := time.Duration(0)
		if okIn && okBrk && okOut {
			liquid = jornada.LiquidDay(in, brk, out)
			weekTotal += liquid
		}
		grid = append(grid, gridDay{
			Key: key, Name: configGridNames[key],
			Status: "Ativo", StatusCls: "bg-secondary-container/40 text-on-secondary-container",
			In: inS, Break: brkS, Out: outS,
			Liquid: jornada.FmtHMShort(liquid),
		})
	}
	grid = append(grid,
		gridDay{Key: "sat", Name: "Sábado", Status: "Compensado", StatusCls: "bg-surface-container-highest text-on-surface-variant", Liquid: "Descanso", Off: true},
		gridDay{Key: "sun", Name: "Domingo", Status: "DSR", StatusCls: "bg-surface-container-highest text-on-surface-variant", Liquid: "Descanso Semanal", Off: true},
	)

	saved := false
	savedMsg := ""
	if m, ok := flash.MessageFromRequest(r); ok && m.Kind == "notice" {
		saved = true
		savedMsg = m.Message
	}

	writeView(w, r, h.views, h.cfg, "app", "configuracoes", amarraData(r, h.site, map[string]any{
		"Title":          "Configurações de Jornada",
		"ActiveNav":      "configuracoes",
		"UserEmail":      h.userEmail(uid),
		"NowClock":       saoPauloNow().Format("15:04:05"),
		"TodayLong":      fmt.Sprintf("%s, %d de %s", ptWeekdays[saoPauloNow().Weekday()], saoPauloNow().Day(), ptMonths[saoPauloNow().Month()-1]),
		"Contract":       pickSetting(get("contract", def.contract), validContracts, def.contract),
		"CompSat":        getBool("comp_sat", def.compSat),
		"Lunch":          pickSetting(get("lunch", def.lunch), validLunch, def.lunch),
		"LunchFlex":      getBool("lunch_flex", def.lunchFlex),
		"BreakExtra":     getBool("break_extra", def.breakExtra),
		"AlertMaster":    getBool("alert_master", def.alertMaster),
		"AlertLead":      pickSetting(get("alert_lead", def.alertLead), validLead, def.alertLead),
		"ChanSound":      getBool("chan_sound", def.chanSound),
		"ChanPush":       getBool("chan_push", def.chanPush),
		"ChanTelegram":   getBool("chan_telegram", def.chanTelegram),
		"AlertTolerance": getBool("alert_tolerance", def.alertToleranc),
		"Cycle":          pickSetting(get("cycle", def.cycle), validCycles, def.cycle),
		"GridDays":       grid,
		"WeekTotal":      jornada.FmtHMShort(weekTotal),
		"Saved":          saved,
		"SavedMsg":       savedMsg,
	}), 0)
}

func (h *ConfiguracoesHandler) Save(w http.ResponseWriter, r *http.Request) {
	uid, ok := session.UserID(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if err := httpx.ParseFormOrJSON(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	def := defaultConfig()
	kv := map[string]string{
		"contract":        pickSetting(r.FormValue("contract"), validContracts, def.contract),
		"lunch":           pickSetting(r.FormValue("lunch"), validLunch, def.lunch),
		"alert_lead":      pickSetting(r.FormValue("alert_lead"), validLead, def.alertLead),
		"cycle":           pickSetting(r.FormValue("cycle"), validCycles, def.cycle),
		"comp_sat":        onOff(r.FormValue("comp_sat")),
		"lunch_flex":      onOff(r.FormValue("lunch_flex")),
		"break_extra":     onOff(r.FormValue("break_extra")),
		"alert_master":    onOff(r.FormValue("alert_master")),
		"chan_sound":      onOff(r.FormValue("chan_sound")),
		"chan_push":       onOff(r.FormValue("chan_push")),
		"chan_telegram":   onOff(r.FormValue("chan_telegram")),
		"alert_tolerance": onOff(r.FormValue("alert_tolerance")),
	}
	for key, d := range configGridDefaults {
		for i, suffix := range []string{"in", "break", "out"} {
			field := "grid_" + key + "_" + suffix
			v := strings.TrimSpace(r.FormValue(field))
			if _, ok := jornada.ParseHHMM(v); ok {
				kv[field] = v
			} else {
				kv[field] = d[i]
			}
		}
	}
	if err := h.store.SetSettings(uid, kv); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flash.Set(w, "notice", "Configurações atualizadas", h.cfg.CookieSecure())
	http.Redirect(w, r, "/configuracoes", http.StatusSeeOther)
}

func (h *ConfiguracoesHandler) Reset(w http.ResponseWriter, r *http.Request) {
	uid, ok := session.UserID(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if err := h.store.DeleteSettings(uid); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	flash.Set(w, "notice", "Padrões da CLT restaurados", h.cfg.CookieSecure())
	http.Redirect(w, r, "/configuracoes", http.StatusSeeOther)
}

func onOff(v string) string {
	if v == "on" || v == "true" || v == "1" {
		return "on"
	}
	return "off"
}

func (h *ConfiguracoesHandler) userEmail(uid int64) string {
	u, err := h.store.FindUserByID(uid)
	if err != nil {
		return ""
	}
	return u.Email
}
