package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/danielxfeng/short-url/apps/backend-chi/internal/api/router"
	"github.com/danielxfeng/short-url/apps/backend-chi/internal/dep"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	// Init the dependency
	d, err := dep.InitDep(slog.LevelDebug)
	if err != nil {
		log.Fatalf("failed to init the dependency, err: %v", err)
	}
	defer dep.CloseDep(d)

	// Init the API server
	r := router.NewRouter(d)
	server := &http.Server{Addr: fmt.Sprintf("0.0.0.0:%d", d.Cfg.Port), Handler: r}

	// Graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			d.Logger.Error("server error", "error", err)
		}
	}()

	<-ctx.Done()

	// Create shutdown context with 30-second timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Trigger graceful shutdown
	if err := server.Shutdown(shutdownCtx); err != nil {
		d.Logger.Error("server forced to shutdown", "error", err)
	}

	d.Logger.Info("server exiting")
}
