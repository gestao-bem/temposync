package handlers

import (
	"net/http"

	"github.com/puppe1990/amarra-cais/pkg/amarra/view"
	"github.com/puppe1990/amarra-cais/pkg/cais"
	"github.com/puppe1990/amarra-cais/pkg/cais/i18n"
	"github.com/puppe1990/amarra-cais/pkg/cais/meta"
)

type HomeHandler struct {
	views   *view.Renderer
	site    meta.Site
	catalog *i18n.Catalog
	cfg     cais.Config
}

func NewHomeHandler(views *view.Renderer, site meta.Site, catalog *i18n.Catalog, cfg cais.Config) *HomeHandler {
	return &HomeHandler{views: views, site: site, catalog: catalog, cfg: cfg}
}

func (h *HomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	view.Write(w, r, h.views, view.Page{
		Layout: "app",
		Name:   "home",
		Data: amarraData(r, h.site, map[string]any{
			"Title":     h.catalog.T("home.title"),
			"ActiveNav": "home",
		}),
	}, h.cfg)
}
