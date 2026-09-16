package app

import "net/http"

// NoStorePages evita que o browser reuse HTML dinâmico do cache heurístico
// (sem Cache-Control o Chrome serve páginas velhas — e scripts inline junto).
// Aplicado só nas rotas dinâmicas: o static registra antes e mantém seu cache.
func NoStorePages(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}
