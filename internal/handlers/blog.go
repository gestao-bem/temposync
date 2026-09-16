package handlers

import (
	"net/http"

	"github.com/puppe1990/amarra-cais/pkg/amarra/view"
	"github.com/puppe1990/amarra-cais/pkg/cais"
	"github.com/puppe1990/amarra-cais/pkg/cais/i18n"
	"github.com/puppe1990/amarra-cais/pkg/cais/meta"
)

type BlogHandler struct {
	views   *view.Renderer
	site    meta.Site
	catalog *i18n.Catalog
	cfg     cais.Config
}

func NewBlogHandler(views *view.Renderer, site meta.Site, catalog *i18n.Catalog, cfg cais.Config) *BlogHandler {
	return &BlogHandler{views: views, site: site, catalog: catalog, cfg: cfg}
}

func (h *BlogHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	view.Write(w, r, h.views, view.Page{
		Layout: "app",
		Name:   "blog",
		Data: amarraData(r, h.site, map[string]any{
			"Title":     "Blog",
			"ActiveNav": "blog",
		}),
	}, h.cfg)
}
