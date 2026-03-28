package router

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/danielxfeng/short-url/apps/backend-chi/internal/api/mymiddleware"
	"github.com/danielxfeng/short-url/apps/backend-chi/internal/api/util"
	"github.com/danielxfeng/short-url/apps/backend-chi/internal/dep"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog/v3"
	"github.com/go-chi/httprate"
)

var logOptions = &httplog.Options{
	Level:         slog.LevelInfo,
	Schema:        httplog.SchemaOTEL,
	RecoverPanics: true,
}

func getCorsOptions(cfg *dep.Config) cors.Options {
	opt := cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}

	if cfg.AppMode == dep.EnvProd {
		opt.AllowedOrigins = []string{cfg.Cors}
	}

	return opt
}

func NewRouter(dep *dep.Dep) http.Handler {
	r := chi.NewRouter()

	r.Use(httplog.RequestLogger(dep.Logger, logOptions)) // It recovers panic
	r.Use(mymiddleware.Helmet())
	// r.Use(middleware.RealIP) enable if behind a proxy
	r.Use(middleware.RequestID)
	r.Use(middleware.CleanPath)

	r.Use(cors.Handler(getCorsOptions(dep.Cfg)))

	r.Use(httprate.LimitByIP(100, 1*time.Minute))
	r.Use(middleware.Compress(5))

	r.Get("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		util.SendJSON(w, 200, map[string]string{"status": "ok"})
	})

	return r
}
