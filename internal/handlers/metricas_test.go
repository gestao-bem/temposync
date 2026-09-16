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

func TestMetricasHandler_RendersHTML(t *testing.T) {
	s := setupTestStore(t)
	h := NewMetricasHandler(setupTestViews(t), s, testSite(), i18n.DefaultCatalog(), cais.Config{})

	uid, err := s.CreateUser("lucas@example.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/metricas", nil)
	req = session.WithUserID(req, uid)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	for _, want := range []string{
		`data-testid="temposync-metricas"`,
		"Análise",
		"Composição do Tempo",
		"Comparativo Histórico Mensal",
	} {
		if !strings.Contains(rr.Body.String(), want) {
			t.Errorf("body missing %q", want)
		}
	}
}

func TestMetricasHandler_redirectsWhenAnonymous(t *testing.T) {
	s := setupTestStore(t)
	h := NewMetricasHandler(setupTestViews(t), s, testSite(), i18n.DefaultCatalog(), cais.Config{})

	req := httptest.NewRequest(http.MethodGet, "/metricas", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", rr.Code)
	}
}

func TestMetricasHandler_ComputesKPIs(t *testing.T) {
	s := setupTestStore(t)
	h := NewMetricasHandler(setupTestViews(t), s, testSite(), i18n.DefaultCatalog(), cais.Config{})

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
		day.Add(8*time.Hour + 2*time.Minute),
		day.Add(12*time.Hour + 1*time.Minute),
		day.Add(13 * time.Hour),
		day.Add(17*time.Hour + 3*time.Minute),
	} {
		if _, err := s.CreatePunch(uid, at, ""); err != nil {
			t.Fatal(err)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/metricas", nil)
	req = session.WithUserID(req, uid)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	body := rr.Body.String()
	for _, want := range []string{"08h 02m", "100.0", "00h 02m"} {
		if !strings.Contains(body, want) {
			t.Errorf("body missing %q", want)
		}
	}
}
