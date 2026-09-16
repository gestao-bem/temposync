package app

import (
	"net/http"

	"github.com/puppe1990/amarra-cais/pkg/amarra/live"
	"github.com/puppe1990/amarra-cais/pkg/cais"
	"github.com/puppe1990/amarra-cais/pkg/cais/middleware"

	"github.com/gestao-bem/temposync/internal/handlers"
)

func registerRoutes(r *cais.Router, deps Deps, cfg cais.Config) {
	home := handlers.NewHomeHandler(deps.Views, deps.Site, deps.Catalog, cfg)
	contact := handlers.NewContactHandler(deps.Views, deps.Store, deps.Site, deps.Catalog, cfg)
	dashboard := handlers.NewDashboardHandler(deps.Views, deps.Store, deps.Site, deps.Catalog, cfg)
	auth := handlers.NewAuthHandler(deps.Views, deps.Store, deps.Site, deps.Store.Sessions(), cfg, deps.Catalog)

	loginLimit := middleware.NewRateLimiter(10, cfg)
	resetLimit := middleware.NewRateLimiter(10, cfg)
	contactLimit := middleware.NewRateLimiter(20, cfg)

	r.Get("/", home.ServeHTTP)
	r.Get("/contact", contact.Get)
	r.Post("/contact", contactLimit.Middleware(http.HandlerFunc(contact.Post)).ServeHTTP)
	r.Get("/login", auth.Login)
	r.Post("/login", loginLimit.Middleware(http.HandlerFunc(auth.LoginPost)).ServeHTTP)
	r.Get("/signup", auth.SignUp)
	r.Post("/signup", loginLimit.Middleware(http.HandlerFunc(auth.SignUpPost)).ServeHTTP)
	r.Get("/forgot-password", auth.ForgotPassword)
	r.Post("/forgot-password", resetLimit.Middleware(http.HandlerFunc(auth.ForgotPasswordPost)).ServeHTTP)
	r.Get("/reset-password", auth.ResetPassword)
	r.Post("/reset-password", resetLimit.Middleware(http.HandlerFunc(auth.ResetPasswordPost)).ServeHTTP)
	r.Post("/logout", auth.LogoutPost)
	r.Post("/locale", handlers.PostLocale(cfg))
	r.Get("/dashboard", middleware.RequireAuthFunc("/login", dashboard.ServeHTTP))
	r.Post("/dashboard/punch", middleware.RequireAuthFunc("/login", dashboard.PunchPost))
	r.Post("/dashboard/punch/delete", middleware.RequireAuthFunc("/login", dashboard.PunchDelete))
	r.Post("/dashboard/mode", middleware.RequireAuthFunc("/login", dashboard.ModePost))

	blog := handlers.NewBlogHandler(deps.Views, deps.Store, deps.Site, deps.Catalog, cfg)
	r.Get("/blog", blog.ServeHTTP)
	r.Get("/blog/{slug}", blog.Post)
	r.Post("/blog/newsletter", blog.NewsletterPost)
	espelho := handlers.NewEspelhoHandler(deps.Views, deps.Store, deps.Site, deps.Catalog, cfg)
	r.Get("/espelho", middleware.RequireAuthFunc("/login", espelho.ServeHTTP))
	r.Get("/espelho/export.csv", middleware.RequireAuthFunc("/login", espelho.ExportCSV))
	r.Get("/espelho/espelho.pdf", middleware.RequireAuthFunc("/login", espelho.ExportPDF))
	r.Post("/espelho/ajustes", middleware.RequireAuthFunc("/login", espelho.AjustePost))
	metricas := handlers.NewMetricasHandler(deps.Views, deps.Store, deps.Site, deps.Catalog, cfg)
	r.Get("/metricas", middleware.RequireAuthFunc("/login", metricas.ServeHTTP))
	bancoHoras := handlers.NewBancoHorasHandler(deps.Views, deps.Store, deps.Site, deps.Catalog, cfg)
	r.Get("/banco-horas", middleware.RequireAuthFunc("/login", bancoHoras.ServeHTTP))
	r.Get("/banco-horas/export.csv", middleware.RequireAuthFunc("/login", bancoHoras.ExportCSV))
	r.Post("/banco-horas/solicitacoes", middleware.RequireAuthFunc("/login", bancoHoras.SolicitacoesPost))
	configuracoes := handlers.NewConfiguracoesHandler(deps.Views, deps.Store, deps.Site, deps.Catalog, cfg)
	r.Get("/configuracoes", middleware.RequireAuthFunc("/login", configuracoes.ServeHTTP))
	r.Post("/configuracoes", middleware.RequireAuthFunc("/login", configuracoes.Save))
	r.Post("/configuracoes/reset", middleware.RequireAuthFunc("/login", configuracoes.Reset))
}

func registerLiveViews(hub *live.Hub, deps Deps) {
	_ = hub
	_ = deps
	// cais:live-views
}
