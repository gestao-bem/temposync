package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/puppe1990/amarra-cais/pkg/cais"
	"github.com/puppe1990/amarra-cais/pkg/cais/flash"
	"github.com/puppe1990/amarra-cais/pkg/cais/i18n"
	"github.com/puppe1990/amarra-cais/pkg/cais/session"

	"github.com/gestao-bem/temposync/internal/store"
)

func newDashboardHandler(t *testing.T) (*DashboardHandler, store.Store) {
	t.Helper()
	s := setupTestStore(t)
	return NewDashboardHandler(setupTestViews(t), s, testSite(), i18n.DefaultCatalog(), cais.Config{}), s
}

func TestDashboardHandler_RendersHTML(t *testing.T) {
	h, s := newDashboardHandler(t)

	uid, err := s.CreateUser("lucas@example.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	req = session.WithUserID(req, uid)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	for _, want := range []string{`data-testid="temposync-dashboard"`, "Previsão de Saída", "Linha do Tempo"} {
		if !strings.Contains(rr.Body.String(), want) {
			t.Errorf("body missing %q", want)
		}
	}
}

func TestDashboardHandler_redirectsWhenAnonymous(t *testing.T) {
	h, _ := newDashboardHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", rr.Code)
	}
}

func TestDashboardHandler_RendersPunches(t *testing.T) {
	h, s := newDashboardHandler(t)

	uid, err := s.CreateUser("lucas@example.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		t.Fatal(err)
	}
	nows := now.In(loc)
	day := time.Date(nows.Year(), nows.Month(), nows.Day(), 0, 0, 0, 0, loc)
	for _, at := range []time.Time{day.Add(8*time.Hour + 32*time.Minute), day.Add(12*time.Hour + 5*time.Minute), day.Add(13*time.Hour + 10*time.Minute)} {
		if _, err := s.CreatePunch(uid, at, ""); err != nil {
			t.Fatal(err)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	req = session.WithUserID(req, uid)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	body := rr.Body.String()
	if !strings.Contains(body, "08:32:00") {
		t.Errorf("missing first punch clock")
	}
	if !strings.Contains(body, "Registrar Saída") {
		t.Errorf("missing next punch label")
	}
}

func TestDashboardHandler_forecastUsesSaoPauloClock(t *testing.T) {
	h, s := newDashboardHandler(t)

	uid, err := s.CreateUser("lucas@example.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		t.Fatal(err)
	}
	nows := time.Now().In(loc)
	day := time.Date(nows.Year(), nows.Month(), nows.Day(), 0, 0, 0, 0, loc)
	entry := day.Add(8*time.Hour + 32*time.Minute)
	if _, err := s.CreatePunch(uid, entry, "entrada"); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	req = session.WithUserID(req, uid)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	body := rr.Body.String()
	// 08:32 + 8h + 1h padrão = 17:32 horário de SP (não 20:32 UTC).
	if !strings.Contains(body, "17:32") {
		t.Errorf("missing SP forecast 17:32")
	}
	if strings.Contains(body, "20:32") {
		t.Errorf("forecast leaked UTC clock")
	}
}

func TestDashboardHandler_includesFlash(t *testing.T) {
	h, s := newDashboardHandler(t)

	uid, err := s.CreateUser("lucas@example.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	req = session.WithUserID(req, uid)
	req = flash.WithMessage(req, flash.Message{Kind: "notice", Message: "Welcome back!"})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Welcome back!") {
		t.Errorf("missing flash notice")
	}
}

func TestDashboard_PunchPost_createsPunch(t *testing.T) {
	h, s := newDashboardHandler(t)

	uid, err := s.CreateUser("lucas@example.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/dashboard/punch", nil)
	req = session.WithUserID(req, uid)
	rr := httptest.NewRecorder()
	h.PunchPost(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303", rr.Code)
	}
	if loc := rr.Header().Get("Location"); loc != "/dashboard" {
		t.Errorf("Location = %q, want /dashboard", loc)
	}
}

func TestDashboard_PunchPost_anonymousRedirects(t *testing.T) {
	h, _ := newDashboardHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/dashboard/punch", nil)
	rr := httptest.NewRecorder()
	h.PunchPost(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", rr.Code)
	}
}
