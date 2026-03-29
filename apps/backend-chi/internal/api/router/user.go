package router

import (
	"net/http"

	"github.com/danielxfeng/short-url/apps/backend-chi/internal/api/mymiddleware"
	"github.com/danielxfeng/short-url/apps/backend-chi/internal/dep"
	"github.com/go-chi/chi/v5"
)

func UserRouter(dep *dep.Dep) http.Handler {
	r := chi.NewRouter()

	r.Route("/auth", func(r chi.Router) {
		r.Get("/{provider}", func(w http.ResponseWriter, r *http.Request) {})
		r.Get("/{provider}/callback", func(w http.ResponseWriter, r *http.Request) {})
	})

	r.Use(mymiddleware.Auth(dep.Cfg.JWTSecret))

	r.Route("/me", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {})
		r.Delete("/", func(w http.ResponseWriter, r *http.Request) {})
	})

	return r
}