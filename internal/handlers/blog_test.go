package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/puppe1990/amarra-cais/pkg/cais"
	"github.com/puppe1990/amarra-cais/pkg/cais/i18n"
)

func TestBlogHandler_RendersHTML(t *testing.T) {
	h := NewBlogHandler(setupTestViews(t), setupTestStore(t), testSite(), i18n.DefaultCatalog(), cais.Config{})

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

func TestBlogHandler_PostBySlug(t *testing.T) {
	h := NewBlogHandler(setupTestViews(t), setupTestStore(t), testSite(), i18n.DefaultCatalog(), cais.Config{})

	req := httptest.NewRequest(http.MethodGet, "/blog/tolerancia-5-minutos-clt-art-58", nil)
	req.SetPathValue("slug", "tolerancia-5-minutos-clt-art-58")
	rr := httptest.NewRecorder()
	h.Post(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	body := rr.Body.String()
	for _, want := range []string{"Tolerância de 5 Minutos", "Artigo 58", "Lucas Silva", "Legislação"} {
		if !strings.Contains(body, want) {
			t.Errorf("body missing %q", want)
		}
	}
}

func TestBlogHandler_PostNotFound(t *testing.T) {
	h := NewBlogHandler(setupTestViews(t), setupTestStore(t), testSite(), i18n.DefaultCatalog(), cais.Config{})

	req := httptest.NewRequest(http.MethodGet, "/blog/nao-existe", nil)
	req.SetPathValue("slug", "nao-existe")
	rr := httptest.NewRecorder()
	h.Post(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "não encontrado") {
		t.Errorf("esperado aviso de não encontrado")
	}
}

func TestBlogHandler_NewsletterPersists(t *testing.T) {
	s := setupTestStore(t)
	h := NewBlogHandler(setupTestViews(t), s, testSite(), i18n.DefaultCatalog(), cais.Config{})

	form := url.Values{"email": {"leitor@example.com"}}
	req := httptest.NewRequest(http.MethodPost, "/blog/newsletter", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.NewsletterPost(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303", rr.Code)
	}
	n, err := s.CountNewsletter()
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("newsletter count = %d, want 1", n)
	}
}

func TestBlogHandler_NewsletterRejectsInvalid(t *testing.T) {
	s := setupTestStore(t)
	h := NewBlogHandler(setupTestViews(t), s, testSite(), i18n.DefaultCatalog(), cais.Config{})

	form := url.Values{"email": {"nao-email"}}
	req := httptest.NewRequest(http.MethodPost, "/blog/newsletter", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.NewsletterPost(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303", rr.Code)
	}
	n, _ := s.CountNewsletter()
	if n != 0 {
		t.Errorf("email inválido não deve persistir: %d", n)
	}
}
