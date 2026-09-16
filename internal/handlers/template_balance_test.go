package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/puppe1990/amarra-cais/pkg/cais"
	"github.com/puppe1990/amarra-cais/pkg/cais/i18n"
	"github.com/puppe1990/amarra-cais/pkg/cais/session"
)

// O Drive morfa #amarra-main cortando o bloco por contagem de <div> fechadas
// (sliceMatchingClose); HTML desbalanceado faz o morph de página inteira
// falhar em silêncio — o clique "não navega" sem erro no console.
func TestPages_BalancedDivTags(t *testing.T) {
	s := setupTestStore(t)
	views := setupTestViews(t)
	site := testSite()
	cfg := cais.Config{}
	cat := i18n.DefaultCatalog()

	uid, err := s.CreateUser("lucas@example.com", "hash")
	if err != nil {
		t.Fatal(err)
	}

	home := NewHomeHandler(views, site, cat, cfg)
	blog := NewBlogHandler(views, site, cat, cfg)
	auth := NewAuthHandler(views, s, site, s.Sessions(), cfg, cat)
	dash := NewDashboardHandler(views, s, site, cat, cfg)
	esp := NewEspelhoHandler(views, s, site, cat, cfg)
	met := NewMetricasHandler(views, s, site, cat, cfg)
	banco := NewBancoHorasHandler(views, s, site, cat, cfg)

	cases := []struct {
		name string
		h    http.HandlerFunc
		auth bool
	}{
		{"home", home.ServeHTTP, false},
		{"blog", blog.ServeHTTP, false},
		{"login", auth.Login, false},
		{"signup", auth.SignUp, false},
		{"dashboard", dash.ServeHTTP, true},
		{"espelho", esp.ServeHTTP, true},
		{"metricas", met.ServeHTTP, true},
		{"banco-horas", banco.ServeHTTP, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.auth {
				req = session.WithUserID(req, uid)
			}
			rr := httptest.NewRecorder()
			tc.h(rr, req)

			body := rr.Body.String()
			opens := strings.Count(body, "<div")
			closes := strings.Count(body, "</div>")
			if opens != closes {
				t.Errorf("<div>=%d </div>=%d (diff %+d): morph do Drive quebra em silêncio", opens, closes, opens-closes)
			}
		})
	}
}
