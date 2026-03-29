package dep

import (
	"log/slog"
	"time"

	db "github.com/danielxfeng/short-url/apps/backend-chi/internal/api/db/sqlc"
	"github.com/getsentry/sentry-go"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Dep struct {
	Cfg    *Config
	Logger *slog.Logger
	DbPool *pgxpool.Pool
}

func NewDep(cfg *Config, logger *slog.Logger, dbPool *pgxpool.Pool) *Dep {
	return &Dep{
		Cfg:    cfg,
		Logger: logger,
		DbPool: dbPool,
	}
}

func InitDep(level slog.Level) (*Dep, error) {
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		return nil, err
	}

	err = InitSentry(cfg.AppMode, cfg.SentryDSN)
	if err != nil {
		return nil, err
	}

	logger := GetLogger(level, cfg.AppMode, cfg.SentryDSN)
	dbPool, err := db.NewPool(cfg.DbURL)
	if err != nil {
		return nil, err
	}
	
	return NewDep(cfg, logger, dbPool), nil
}

func CloseDep(dep *Dep) {
	sentry.Flush(2 * time.Second)
	db.ClosePool(dep.DbPool)
}
