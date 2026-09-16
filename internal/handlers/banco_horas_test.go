package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/puppe1990/amarra-cais/pkg/cais"
	"github.com/puppe1990/amarra-cais/pkg/cais/i18n"
	"github.com/puppe1990/amarra-cais/pkg/cais/session"

	"github.com/gestao-bem/temposync/internal/store"
)

func newBancoHorasHandler(t *testing.T) (*BancoHorasHandler, store.Store) {
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

func TestBancoHoras_SolicitaFolgaEAprova(t *testing.T) {
	h, s := newBancoHorasHandler(t)

	uid, err := s.CreateUser("lucas@example.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	loc, _ := time.LoadLocation("America/Sao_Paulo")
	nows := time.Now().In(loc)
	// saldo de +2h hoje para cobrir a folga de 2h
	day := time.Date(nows.Year(), nows.Month(), nows.Day(), 0, 0, 0, 0, loc)
	for _, at := range []time.Time{day.Add(8 * time.Hour), day.Add(12 * time.Hour), day.Add(13 * time.Hour), day.Add(19 * time.Hour)} {
		if _, err := s.CreatePunch(uid, at, ""); err != nil {
			t.Fatal(err)
		}
	}

	form := url.Values{
		"acao":   {"criar"},
		"kind":   {"folga"},
		"day":    {nows.AddDate(0, 0, 3).Format("2006-01-02")},
		"hours":  {"2"},
		"reason": {"compensar plantão"},
	}
	req := httptest.NewRequest(http.MethodPost, "/banco-horas/solicitacoes", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = session.WithUserID(req, uid)
	rr := httptest.NewRecorder()
	h.SolicitacoesPost(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("criar: status = %d, want 303", rr.Code)
	}

	list, err := s.ListRequests(uid)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].Kind != "folga" || list[0].Minutes != 120 || list[0].Status != "pendente" {
		t.Fatalf("requests = %+v", list)
	}

	// aprovar
	form = url.Values{"acao": {"aprovar"}, "id": {strconv.FormatInt(list[0].ID, 10)}}
	req = httptest.NewRequest(http.MethodPost, "/banco-horas/solicitacoes", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = session.WithUserID(req, uid)
	rr = httptest.NewRecorder()
	h.SolicitacoesPost(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("aprovar: status = %d, want 303", rr.Code)
	}

	list, _ = s.ListRequests(uid)
	if list[0].Status != "aprovada" {
		t.Fatalf("status = %s, want aprovada", list[0].Status)
	}
	// folga virou punch kind=folga com 120min
	folgaDay := nows.AddDate(0, 0, 3)
	punches, err := s.ListPunches(uid, dayStart(folgaDay), dayStart(folgaDay).Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(punches) != 1 || punches[0].Kind != "folga" || punches[0].Minutes != 120 {
		t.Fatalf("folga punch = %+v", punches)
	}
}

func TestBancoHoras_CancelarNaoCriaFolga(t *testing.T) {
	h, s := newBancoHorasHandler(t)

	uid, err := s.CreateUser("lucas@example.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	loc, _ := time.LoadLocation("America/Sao_Paulo")
	day := saoPauloNow().AddDate(0, 0, 2)
	_ = loc

	form := url.Values{"acao": {"criar"}, "kind": {"folga"}, "day": {day.Format("2006-01-02")}, "hours": {"4"}, "reason": {"x"}}
	req := httptest.NewRequest(http.MethodPost, "/banco-horas/solicitacoes", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = session.WithUserID(req, uid)
	h.SolicitacoesPost(httptest.NewRecorder(), req)

	list, _ := s.ListRequests(uid)
	form = url.Values{"acao": {"cancelar"}, "id": {strconv.FormatInt(list[0].ID, 10)}}
	req = httptest.NewRequest(http.MethodPost, "/banco-horas/solicitacoes", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = session.WithUserID(req, uid)
	h.SolicitacoesPost(httptest.NewRecorder(), req)

	list, _ = s.ListRequests(uid)
	if list[0].Status != "cancelada" {
		t.Fatalf("status = %s", list[0].Status)
	}
	punches, _ := s.ListPunches(uid, dayStart(day), dayStart(day).Add(24*time.Hour))
	if len(punches) != 0 {
		t.Errorf("cancelar não pode criar folga: %+v", punches)
	}
}
