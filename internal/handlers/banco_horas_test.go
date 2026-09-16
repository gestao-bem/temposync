package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/puppe1990/amarra-cais/pkg/cais"
	"github.com/puppe1990/amarra-cais/pkg/cais/i18n"
	"github.com/puppe1990/amarra-cais/pkg/cais/session"
)

func newBancoHorasHandler(t *testing.T) (*BancoHorasHandler, interface {
	CreateUser(string, string) (int64, error)
	CreatePunch(int64, time.Time, string) (int64, error)
}) {
	t.Helper()
	s := setupTestStore(t)
	return NewBancoHorasHandler(setupTestViews(t), s, testSite(), i18n.DefaultCatalog(), cais.Config{}), s
}

func TestBancoHorasHandler_RendersHTML(t *testing.T) {
	h, s := newBancoHorasHandler(t)

	uid, err := s.CreateUser("lucas@example.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/banco-horas", nil)
	req = session.WithUserID(req, uid)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	for _, want := range []string{
		`data-testid="temposync-banco-horas"`,
		"Banco de Horas",
		"Extrato de Lançamentos",
		"Simulador de Folga",
	} {
		if !strings.Contains(rr.Body.String(), want) {
			t.Errorf("body missing %q", want)
		}
	}
}

func TestBancoHorasHandler_redirectsWhenAnonymous(t *testing.T) {
	h, _ := newBancoHorasHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/banco-horas", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", rr.Code)
	}
}

func TestBancoHorasHandler_ComputesSaldoAndExtrato(t *testing.T) {
	h, s := newBancoHorasHandler(t)

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
	// jornada de 10h → +02h 00m efetivos
	for _, at := range []time.Time{
		day.Add(8 * time.Hour),
		day.Add(12 * time.Hour),
		day.Add(13 * time.Hour),
		day.Add(19 * time.Hour),
	} {
		if _, err := s.CreatePunch(uid, at, ""); err != nil {
			t.Fatal(err)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/banco-horas", nil)
	req = session.WithUserID(req, uid)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	body := rr.Body.String()
	// nota: Go escapa "+" como &#43; no HTML (o navegador exibe "+").
	for _, want := range []string{"02h 00m", "Jornada excedente"} {
		if !strings.Contains(body, want) {
			t.Errorf("body missing %q", want)
		}
	}
}

func TestBancoHorasHandler_FilterCredOnly(t *testing.T) {
	h, s := newBancoHorasHandler(t)

	uid, err := s.CreateUser("lucas@example.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		t.Fatal(err)
	}
	nows := time.Now().In(loc)
	// ontem: +2h; hoje: -3h (compensação)
	yesterday := time.Date(nows.Year(), nows.Month(), nows.Day()-1, 0, 0, 0, 0, loc)
	today := time.Date(nows.Year(), nows.Month(), nows.Day(), 0, 0, 0, 0, loc)
	for _, at := range []time.Time{
		yesterday.Add(8 * time.Hour), yesterday.Add(12 * time.Hour),
		yesterday.Add(13 * time.Hour), yesterday.Add(18 * time.Hour),
	} {
		if _, err := s.CreatePunch(uid, at, ""); err != nil {
			t.Fatal(err)
		}
	}
	for _, at := range []time.Time{
		today.Add(8 * time.Hour), today.Add(12 * time.Hour), today.Add(13 * time.Hour),
	} {
		if _, err := s.CreatePunch(uid, at, ""); err != nil {
			t.Fatal(err)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/banco-horas?tipo=cred", nil)
	req = session.WithUserID(req, uid)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Jornada excedente") {
		t.Errorf("credito must render")
	}
	if strings.Contains(body, "Folga / compensação de") {
		t.Errorf("filtro cred não deve renderizar linhas de compensação")
	}
}

func TestBancoHorasHandler_EmptyState(t *testing.T) {
	h, s := newBancoHorasHandler(t)

	uid, err := s.CreateUser("lucas@example.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/banco-horas", nil)
	req = session.WithUserID(req, uid)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Nenhum lançamento") {
		t.Errorf("sem batidas deve mostrar vazio")
	}
}
