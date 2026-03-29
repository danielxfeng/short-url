package router

import (
	"net/http"

	"github.com/danielxfeng/short-url/apps/backend-chi/internal/api/mymiddleware"
	"github.com/danielxfeng/short-url/apps/backend-chi/internal/dep"
	"github.com/go-chi/chi/v5"
)

func ShortURLRouter(dep *dep.Dep) http.Handler {
	r := chi.NewRouter()

	r.Get("/{code}", func(w http.ResponseWriter, r *http.Request) {
		// Handle short URL logic here
	})

	r.Use(mymiddleware.Auth(dep.Cfg.JWTSecret))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		// Handle listing URLs logic here
	})

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		// Handle URL shortening logic here
	})

	r.Delete("/{code}", func(w http.ResponseWriter, r *http.Request) {
		// Handle URL deletion logic here
	})

	return r
}