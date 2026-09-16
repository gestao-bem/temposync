package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/puppe1990/amarra-cais/pkg/cais"
	"github.com/puppe1990/amarra-cais/pkg/cais/flash"
	"github.com/puppe1990/amarra-cais/pkg/cais/i18n"
	"github.com/puppe1990/amarra-cais/pkg/cais/session"

	"github.com/gestao-bem/temposync/internal/store"
)

func newConfiguracoesHandler(t *testing.T) (*ConfiguracoesHandler, store.Store) {
	t.Helper()
	s := setupTestStore(t)
	return NewConfiguracoesHandler(setupTestViews(t), s, testSite(), i18n.DefaultCatalog(), cais.Config{}), s
}

func TestConfiguracoesHandler_RendersHTML(t *testing.T) {
	h, s := newConfiguracoesHandler(t)

	uid, err := s.CreateUser("lucas@example.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/configuracoes", nil)
	req = session.WithUserID(req, uid)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	for _, want := range []string{
		`data-testid="temposync-configuracoes"`,
		"Configurações de Jornada",
		"Modelo de Contrato",
		"Regras de Intervalo",
		"Salvar Alterações",
		`action="/configuracoes"`,
		`name="csrf_token"`,
	} {
		if !strings.Contains(rr.Body.String(), want) {
			t.Errorf("body missing %q", want)
		}
	}
}

func TestConfiguracoesHandler_redirectsWhenAnonymous(t *testing.T) {
	h, _ := newConfiguracoesHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/configuracoes", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", rr.Code)
	}
}

func TestConfiguracoesHandler_SavePersistsSettings(t *testing.T) {
	h, s := newConfiguracoesHandler(t)

	uid, err := s.CreateUser("lucas@example.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	form := url.Values{
		"contract":     {"clt-40"},
		"lunch":        {"75"},
		"comp_sat":     {"on"},
		"alert_lead":   {"30"},
		"cycle":        {"anual"},
		"grid_mon_in":  {"09:00"},
		"grid_mon_out": {"18:00"},
		"grid_mon_brk": {"01:00"},
	}
	req := httptest.NewRequest(http.MethodPost, "/configuracoes", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = session.WithUserID(req, uid)
	rr := httptest.NewRecorder()
	h.Save(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303, body: %s", rr.Code, rr.Body.String())
	}
	if loc := rr.Header().Get("Location"); loc != "/configuracoes" {
		t.Errorf("Location = %q", loc)
	}

	got, err := s.GetSettings(uid)
	if err != nil {
		t.Fatal(err)
	}
	if got["contract"] != "clt-40" || got["lunch"] != "75" || got["alert_lead"] != "30" || got["cycle"] != "anual" {
		t.Errorf("settings = %v", got)
	}
	if got["grid_mon_in"] != "09:00" || got["grid_mon_out"] != "18:00" {
		t.Errorf("escala = %v", got)
	}
	if _, ok := got["comp_sat"]; !ok {
		t.Errorf("comp_sat ausente: %v", got)
	}
}

func TestConfiguracoesHandler_RejectsInvalidValues(t *testing.T) {
	h, s := newConfiguracoesHandler(t)

	uid, err := s.CreateUser("lucas@example.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	form := url.Values{
		"contract":     {"contrato-invalido"},
		"lunch":        {"999"},
		"alert_lead":   {"99"},
		"cycle":        {"diario"},
		"grid_mon_in":  {"25:99"},
		"grid_mon_out": {"abc"},
		"grid_mon_brk": {"01:00"},
	}
	req := httptest.NewRequest(http.MethodPost, "/configuracoes", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = session.WithUserID(req, uid)
	rr := httptest.NewRecorder()
	h.Save(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303", rr.Code)
	}
	got, err := s.GetSettings(uid)
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range got {
		if k == "contract" && v == "contrato-invalido" {
			t.Errorf("valor inválido persistido: %v", got)
		}
		if k == "lunch" && v == "999" {
			t.Errorf("valor inválido persistido: %v", got)
		}
		if k == "grid_mon_in" && v == "25:99" {
			t.Errorf("horário inválido persistido: %v", got)
		}
	}
}

func TestConfiguracoesHandler_ResetDeletesSettings(t *testing.T) {
	h, s := newConfiguracoesHandler(t)

	uid, err := s.CreateUser("lucas@example.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SetSettings(uid, map[string]string{"contract": "pj"}); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/configuracoes/reset", nil)
	req = session.WithUserID(req, uid)
	rr := httptest.NewRecorder()
	h.Reset(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303", rr.Code)
	}
	got, _ := s.GetSettings(uid)
	if len(got) != 0 {
		t.Errorf("reset não limpou: %v", got)
	}
}

func TestConfiguracoesHandler_ShowsSavedToast(t *testing.T) {
	h, s := newConfiguracoesHandler(t)

	uid, err := s.CreateUser("lucas@example.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/configuracoes", nil)
	req = session.WithUserID(req, uid)
	req = flash.WithMessage(req, flash.Message{Kind: "notice", Message: "Configurações atualizadas"})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if !strings.Contains(rr.Body.String(), "Configurações Atualizadas") {
		t.Errorf("toast de sucesso esperado")
	}
	if !strings.Contains(rr.Body.String(), "opacity-100") {
		t.Errorf("toast deve iniciar visível quando há flash notice")
	}
}
