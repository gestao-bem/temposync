package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/puppe1990/amarra-cais/pkg/cais"
	"github.com/puppe1990/amarra-cais/pkg/cais/i18n"

	"github.com/gestao-bem/temposync/internal/store"
)

func newAuthHandlerForSignup(t *testing.T) (*AuthHandler, store.Store) {
	t.Helper()
	s, err := store.NewSQLiteStore(":memory:", "development")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	h := NewAuthHandler(setupTestViews(t), s, testSite(), s.Sessions(), cais.Config{}, i18n.DefaultCatalog())
	return h, s
}

func TestAuth_SignUpPost_createsUserAndRedirects(t *testing.T) {
	h, s := newAuthHandlerForSignup(t)

	form := url.Values{}
	form.Set("email", "signup@example.com")
	form.Set("password", "password123")
	form.Set("password_confirmation", "password123")
	req := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.SignUpPost(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303, body: %s", rr.Code, rr.Body.String())
	}
	if rr.Header().Get("Location") != "/dashboard" {
		t.Errorf("Location = %q, want /dashboard", rr.Header().Get("Location"))
	}

	user, err := s.FindUserByEmail("signup@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if user.ID == 0 {
		t.Fatal("user id = 0")
	}
}

func TestAuth_SignUpPost_duplicateEmail_returnsError(t *testing.T) {
	h, _ := newAuthHandlerForSignup(t)

	form := url.Values{}
	form.Set("email", "signup@example.com")
	form.Set("password", "password123")
	form.Set("password_confirmation", "password123")
	req := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.SignUpPost(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("first signup status = %d, want 303", rr.Code)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(form.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr2 := httptest.NewRecorder()
	h.SignUpPost(rr2, req2)
	if rr2.Code != http.StatusUnprocessableEntity {
		t.Fatalf("duplicate signup status = %d, want 422", rr2.Code)
	}
	if !strings.Contains(rr2.Body.String(), "already registered") {
		t.Errorf("missing email taken error, got: %s", rr2.Body.String())
	}
}

func TestAuth_SignUp_RendersForm(t *testing.T) {
	h, _ := newAuthHandlerForSignup(t)

	req := httptest.NewRequest(http.MethodGet, "/signup", nil)
	rr := httptest.NewRecorder()
	h.SignUp(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	body := rr.Body.String()
	for _, want := range []string{
		`data-testid="temposync-login"`,
		`data-signup-mode="true"`,
		`action="/signup"`,
		"Comece a controlar sua jornada gratuitamente",
		"Criar Minha Conta Grátis",
		"Já possui conta? Fazer Login",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("body missing %q", want)
		}
	}
	// SSR do título: o h2 não pode trazer o texto do modo login (o JS contém
	// ambos os textos para a troca de abas, por isso checamos o bloco do h2).
	start := strings.Index(body, `id="formTitle"`)
	if start == -1 {
		t.Fatal("formTitle not found")
	}
	h2 := body[start:]
	end := strings.Index(h2, "</h2>")
	if end == -1 {
		t.Fatal("formTitle not closed")
	}
	h2 = h2[:end]
	if !strings.Contains(h2, "Comece a controlar sua jornada gratuitamente") {
		t.Errorf("formTitle not in signup mode: %s", h2)
	}
	if strings.Contains(h2, "Bem-vindo de volta") {
		t.Errorf("formTitle leaked login mode: %s", h2)
	}
}
