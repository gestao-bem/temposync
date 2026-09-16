package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/puppe1990/amarra-cais/pkg/cais"
	"github.com/puppe1990/amarra-cais/pkg/cais/flash"
	"github.com/puppe1990/amarra-cais/pkg/cais/i18n"
)

func TestDashboardHandler_RendersHTML(t *testing.T) {
	h := NewDashboardHandler(setupTestViews(t), setupTestStore(t), testSite(), i18n.DefaultCatalog(), cais.Config{})

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if !strings.Contains(rr.Body.String(), "Dashboard") {
		t.Errorf("missing dashboard heading, got: %s", rr.Body.String())
	}
}

func TestDashboardHandler_includesFlash(t *testing.T) {
	h := NewDashboardHandler(setupTestViews(t), setupTestStore(t), testSite(), i18n.DefaultCatalog(), cais.Config{})

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	req = flash.WithMessage(req, flash.Message{Kind: "notice", Message: "Welcome back!"})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Welcome back!") {
		t.Errorf("missing flash notice, got: %s", rr.Body.String())
	}
}
