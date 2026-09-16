package handlers

import (
	"net/http"

	"github.com/puppe1990/amarra-cais/pkg/amarra/view"
	"github.com/puppe1990/amarra-cais/pkg/cais"
	"github.com/puppe1990/amarra-cais/pkg/cais/i18n"
	"github.com/puppe1990/amarra-cais/pkg/cais/meta"

	"github.com/gestao-bem/temposync/internal/store"
)

type DashboardHandler struct {
	views   *view.Renderer
	store   store.Store
	site    meta.Site
	catalog *i18n.Catalog
	cfg     cais.Config
}

func NewDashboardHandler(views *view.Renderer, s store.Store, site meta.Site, catalog *i18n.Catalog, cfg cais.Config) *DashboardHandler {
	return &DashboardHandler{views: views, store: s, site: site, catalog: catalog, cfg: cfg}
}

func (h *DashboardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	count, err := h.store.CountContacts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeView(w, r, h.views, h.cfg, "app", "dashboard", amarraData(r, h.site, map[string]any{
		"Title":         h.catalog.T("dashboard.title"),
		"ActiveNav":     "dashboard",
		"TotalContacts": count,
		"Env":           h.cfg.Env,
	}), 0)
}
