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

func TestEspelho_ExportCSV(t *testing.T) {
	s := setupTestStore(t)
	h := NewEspelhoHandler(setupTestViews(t), s, testSite(), i18n.DefaultCatalog(), cais.Config{})

	uid, err := s.CreateUser("lucas@example.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	loc, _ := time.LoadLocation("America/Sao_Paulo")
	nows := time.Now().In(loc)
	day := time.Date(nows.Year(), nows.Month(), nows.Day(), 0, 0, 0, 0, loc)
	for _, at := range []time.Time{day.Add(8 * time.Hour), day.Add(12 * time.Hour), day.Add(13 * time.Hour), day.Add(17 * time.Hour)} {
		if _, err := s.CreatePunch(uid, at, ""); err != nil {
			t.Fatal(err)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/espelho/export.csv", nil)
	req = session.WithUserID(req, uid)
	rr := httptest.NewRecorder()
	h.ExportCSV(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, "text/csv") {
		t.Errorf("Content-Type = %q", ct)
	}
	if cd := rr.Header().Get("Content-Disposition"); !strings.Contains(cd, "attachment") || !strings.Contains(cd, ".csv") {
		t.Errorf("Content-Disposition = %q", cd)
	}
	body := rr.Body.String()
	if !strings.HasPrefix(body, "\ufeff") {
		t.Errorf("CSV deve ter BOM UTF-8 para Excel pt-BR")
	}
	for _, want := range []string{"Data", "Entrada 1", "08:00", "17:00", "08h 00m"} {
		if !strings.Contains(body, want) {
			t.Errorf("csv missing %q", want)
		}
	}
	if !strings.Contains(body, ";") {
		t.Errorf("separador pt-BR deve ser ;")
	}
}

func TestBancoHoras_ExportCSV(t *testing.T) {
	s := setupTestStore(t)
	h := NewBancoHorasHandler(setupTestViews(t), s, testSite(), i18n.DefaultCatalog(), cais.Config{})

	uid, err := s.CreateUser("lucas@example.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	loc, _ := time.LoadLocation("America/Sao_Paulo")
	nows := time.Now().In(loc)
	day := time.Date(nows.Year(), nows.Month(), nows.Day(), 0, 0, 0, 0, loc)
	for _, at := range []time.Time{day.Add(8 * time.Hour), day.Add(12 * time.Hour), day.Add(13 * time.Hour), day.Add(19 * time.Hour)} {
		if _, err := s.CreatePunch(uid, at, ""); err != nil {
			t.Fatal(err)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/banco-horas/export.csv", nil)
	req = session.WithUserID(req, uid)
	rr := httptest.NewRecorder()
	h.ExportCSV(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	body := rr.Body.String()
	for _, want := range []string{"Data", "Descrição", "Variação", "Saldo Parcial", "Jornada excedente"} {
		if !strings.Contains(body, want) {
			t.Errorf("csv missing %q", want)
		}
	}
}

func TestExports_redirectWhenAnonymous(t *testing.T) {
	s := setupTestStore(t)
	esp := NewEspelhoHandler(setupTestViews(t), s, testSite(), i18n.DefaultCatalog(), cais.Config{})
	ban := NewBancoHorasHandler(setupTestViews(t), s, testSite(), i18n.DefaultCatalog(), cais.Config{})

	for name, fn := range map[string]http.HandlerFunc{"espelho": esp.ExportCSV, "banco": ban.ExportCSV} {
		req := httptest.NewRequest(http.MethodGet, "/x.csv", nil)
		rr := httptest.NewRecorder()
		fn(rr, req)
		if rr.Code != http.StatusSeeOther {
			t.Errorf("%s: status = %d, want 303", name, rr.Code)
		}
	}
}
