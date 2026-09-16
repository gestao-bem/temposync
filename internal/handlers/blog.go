package handlers

import (
	"net/http"
	"strings"

	"github.com/puppe1990/amarra-cais/pkg/amarra/view"
	"github.com/puppe1990/amarra-cais/pkg/cais"
	"github.com/puppe1990/amarra-cais/pkg/cais/flash"
	"github.com/puppe1990/amarra-cais/pkg/cais/httpx"
	"github.com/puppe1990/amarra-cais/pkg/cais/i18n"
	"github.com/puppe1990/amarra-cais/pkg/cais/meta"
	"github.com/puppe1990/amarra-cais/pkg/cais/validate"

	"github.com/gestao-bem/temposync/internal/store"
)

type BlogHandler struct {
	views   *view.Renderer
	store   store.Store
	site    meta.Site
	catalog *i18n.Catalog
	cfg     cais.Config
}

func NewBlogHandler(views *view.Renderer, s store.Store, site meta.Site, catalog *i18n.Catalog, cfg cais.Config) *BlogHandler {
	return &BlogHandler{views: views, store: s, site: site, catalog: catalog, cfg: cfg}
}

func (h *BlogHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	writeView(w, r, h.views, h.cfg, "app", "blog", amarraData(r, h.site, map[string]any{
		"Title":     "Blog",
		"ActiveNav": "blog",
	}), 0)
}

// Post renderiza o artigo por slug (conteúdo real do banco).
func (h *BlogHandler) Post(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimSpace(r.PathValue("slug"))
	post, err := h.store.PostBySlug(slug)
	if err != nil {
		writeView(w, r, h.views, h.cfg, "app", "blog_post", amarraData(r, h.site, map[string]any{
			"Title":     "Artigo não encontrado",
			"ActiveNav": "blog",
			"NotFound":  true,
			"Slug":      slug,
		}), http.StatusNotFound)
		return
	}
	paragraphs := splitParagraphs(post.Body)
	writeView(w, r, h.views, h.cfg, "app", "blog_post", amarraData(r, h.site, map[string]any{
		"Title":      post.Title,
		"ActiveNav":  "blog",
		"Post":       post,
		"Paragraphs": paragraphs,
		"Category":   blogCategoryLabel(post.Category),
	}), 0)
}

func splitParagraphs(body string) []string {
	raw := strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n\n")
	out := make([]string, 0, len(raw))
	for _, p := range raw {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

var blogCategoryLabels = map[string]string{
	"legislacao":    "Legislação & CLT",
	"produtividade": "Produtividade & Foco",
	"gestao":        "Gestão de Ponto & RH",
	"saude":         "Saúde Laboral & Burnout",
}

func blogCategoryLabel(key string) string {
	if l, ok := blogCategoryLabels[key]; ok {
		return l
	}
	return "Artigos"
}

// NewsletterPost valida e persiste o email (idempotente).
func (h *BlogHandler) NewsletterPost(w http.ResponseWriter, r *http.Request) {
	if err := httpx.ParseFormOrJSON(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	email := strings.TrimSpace(r.FormValue("email"))
	if err := validate.Email(email); err != nil {
		flash.Set(w, "error", "Informe um email válido para assinar", h.cfg.CookieSecure())
		http.Redirect(w, r, "/blog", http.StatusSeeOther)
		return
	}
	added, err := h.store.SubscribeNewsletter(email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if added {
		flash.Set(w, "notice", "Inscrição realizada! Verifique sua caixa de entrada.", h.cfg.CookieSecure())
	} else {
		flash.Set(w, "notice", "Este email já está na nossa lista. Obrigado!", h.cfg.CookieSecure())
	}
	http.Redirect(w, r, "/blog", http.StatusSeeOther)
}
