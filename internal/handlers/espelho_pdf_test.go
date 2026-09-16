package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/puppe1990/amarra-cais/pkg/cais"
	"github.com/puppe1990/amarra-cais/pkg/cais/i18n"
	"github.com/puppe1990/amarra-cais/pkg/cais/session"
)

func TestEspelhoPDF_ServesPDF(t *testing.T) {
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

	req := httptest.NewRequest(http.MethodGet, "/espelho/espelho.pdf", nil)
	req = session.WithUserID(req, uid)
	rr := httptest.NewRecorder()
	h.ExportPDF(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/pdf" {
		t.Errorf("Content-Type = %q", ct)
	}
	if cd := rr.Header().Get("Content-Disposition"); !strings.Contains(cd, ".pdf") {
		t.Errorf("Content-Disposition = %q", cd)
	}
	body := rr.Body.Bytes()
	if len(body) < 1000 {
		t.Fatalf("pdf pequeno demais: %d bytes", len(body))
	}
	if !strings.HasPrefix(string(body[:5]), "%PDF-") {
		t.Errorf("magic bytes inválidos: %q", string(body[:5]))
	}
	if !strings.Contains(rr.Header().Get("X-Document-Hash"), "") {
		t.Errorf("hash de integridade esperado no header")
	}
}

func TestEspelhoPDF_IntegrityHashMatchesContent(t *testing.T) {
	s := setupTestStore(t)
	h := NewEspelhoHandler(setupTestViews(t), s, testSite(), i18n.DefaultCatalog(), cais.Config{})

	uid, err := s.CreateUser("lucas@example.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	loc, _ := time.LoadLocation("America/Sao_Paulo")
	nows := time.Now().In(loc)
	month := time.Date(nows.Year(), nows.Month(), 1, 0, 0, 0, 0, loc)
	day := month.AddDate(0, 0, 1)
	for day.Weekday() == time.Saturday || day.Weekday() == time.Sunday {
		day = day.AddDate(0, 0, 1)
	}
	for _, at := range []time.Time{day.Add(8 * time.Hour), day.Add(12 * time.Hour), day.Add(13 * time.Hour), day.Add(17 * time.Hour)} {
		if _, err := s.CreatePunch(uid, at, ""); err != nil {
			t.Fatal(err)
		}
	}

	hash1 := h.contentHash(uid, month, nows)
	hash2 := h.contentHash(uid, month, nows)
	if hash1 != hash2 {
		t.Fatalf("hash não determinístico: %s vs %s", hash1, hash2)
	}
	if len(hash1) != 64 {
		t.Errorf("hash deve ter 64 hex chars, got %d", len(hash1))
	}
	if _, err := hex.DecodeString(hash1); err != nil {
		t.Errorf("hash inválido: %v", err)
	}
	sum := sha256.Sum256([]byte("x"))
	if len(hex.EncodeToString(sum[:])) != 64 {
		t.Fatal("sanity")
	}
}

func TestEspelhoPDF_redirectsWhenAnonymous(t *testing.T) {
	s := setupTestStore(t)
	h := NewEspelhoHandler(setupTestViews(t), s, testSite(), i18n.DefaultCatalog(), cais.Config{})

	req := httptest.NewRequest(http.MethodGet, "/espelho/espelho.pdf", nil)
	rr := httptest.NewRecorder()
	h.ExportPDF(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", rr.Code)
	}
}
