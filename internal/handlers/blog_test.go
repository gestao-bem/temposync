package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/puppe1990/amarra-cais/pkg/cais"

	appi18n "github.com/gestao-bem/temposync/internal/i18n"
)

func TestBlogHandler_RendersHTML(t *testing.T) {
	h := NewBlogHandler(setupTestViews(t), testSite(), appi18n.DefaultCatalog(), cais.Config{})

	req := httptest.NewRequest(http.MethodGet, "/blog", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	body := rr.Body.String()
	for _, want := range []string{
		`data-testid="temposync-blog"`,
		"Hub de Conhecimento",
		"Tolerância de 5 Minutos",
		"Calculadoras Gratuitas",
		"Assinar Newsletter",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("body missing %q", want)
		}
	}
}
