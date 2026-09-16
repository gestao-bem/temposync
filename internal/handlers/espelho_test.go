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

func TestEspelhoHandler_RendersHTML(t *testing.T) {
	s := setupTestStore(t)
	h := NewEspelhoHandler(setupTestViews(t), s, testSite(), i18n.DefaultCatalog(), cais.Config{})

	uid, err := s.CreateUser("lucas@example.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/espelho", nil)
	req = session.WithUserID(req, uid)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	for _, want := range []string{
		`data-testid="temposync-espelho"`,
		"Espelho de Ponto",
		"Registros Diários da Folha",
		"Fechamento",
	} {
		if !strings.Contains(rr.Body.String(), want) {
			t.Errorf("body missing %q", want)
		}
	}
}

func TestEspelhoHandler_RendersPunches(t *testing.T) {
	s := setupTestStore(t)
	h := NewEspelhoHandler(setupTestViews(t), s, testSite(), i18n.DefaultCatalog(), cais.Config{})

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
	for _, at := range []time.Time{
		day.Add(7*time.Hour + 58*time.Minute),
		day.Add(12*time.Hour + 4*time.Minute),
		day.Add(13*time.Hour + 2*time.Minute),
		day.Add(17*time.Hour + 45*time.Minute),
	} {
		if _, err := s.CreatePunch(uid, at, ""); err != nil {
			t.Fatal(err)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/espelho", nil)
	req = session.WithUserID(req, uid)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	body := rr.Body.String()
	// nota: Go escapa "+" como &#43; no HTML (o navegador exibe "+").
	for _, want := range []string{"07:58", "17:45", "08h 49m", "00h 49m"} {
		if !strings.Contains(body, want) {
			t.Errorf("body missing %q", want)
		}
	}
}

func TestEspelhoHandler_redirectsWhenAnonymous(t *testing.T) {
	s := setupTestStore(t)
	h := NewEspelhoHandler(setupTestViews(t), s, testSite(), i18n.DefaultCatalog(), cais.Config{})

	req := httptest.NewRequest(http.MethodGet, "/espelho", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", rr.Code)
	}
}
