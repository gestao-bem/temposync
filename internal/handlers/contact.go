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

	"github.com/gestao-bem/temposync/internal/models"
	"github.com/gestao-bem/temposync/internal/store"
)

type ContactHandler struct {
	views   *view.Renderer
	store   store.Store
	site    meta.Site
	catalog *i18n.Catalog
	cfg     cais.Config
}

func NewContactHandler(views *view.Renderer, s store.Store, site meta.Site, catalog *i18n.Catalog, cfg cais.Config) *ContactHandler {
	return &ContactHandler{views: views, store: s, site: site, catalog: catalog, cfg: cfg}
}

func (h *ContactHandler) Get(w http.ResponseWriter, r *http.Request) {
	writeView(w, r, h.views, h.cfg, "app", "contact", amarraData(r, h.site, map[string]any{
		"Title":     h.catalog.T("contact.title"),
		"ActiveNav": "contact",
	}), 0)
}

func (h *ContactHandler) Post(w http.ResponseWriter, r *http.Request) {
	if err := httpx.ParseFormOrJSON(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	email := strings.TrimSpace(r.FormValue("email"))

	var errs validate.FieldErrors
	if name == "" {
		errs.Add("name", h.catalog.T("contact.name_required"))
	}
	if err := validate.Email(email); err != nil {
		msg := h.catalog.T("contact.email_required")
		if email != "" {
			msg = h.catalog.T("contact.email_invalid")
		}
		errs.Add("email", msg)
	}
	if errs.Any() {
		writeView(w, r, h.views, h.cfg, "app", "contact", amarraData(r, h.site, map[string]any{
			"Title":     h.catalog.T("contact.title"),
			"ActiveNav": "contact",
			"Errors":    errs,
			"Name":      name,
			"Email":     email,
		}), http.StatusUnprocessableEntity)
		return
	}

	if _, err := h.store.InsertContact(models.Contact{Name: name, Email: email}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	flash.Set(w, "success", "Message sent successfully.", h.cfg.CookieSecure())
	http.Redirect(w, r, "/contact", http.StatusSeeOther)
}
