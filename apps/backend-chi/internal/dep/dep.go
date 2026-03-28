package dep

import (
	"log/slog"
	"time"

	"github.com/getsentry/sentry-go"
)

type Dep struct {
	Cfg    *Config
	Logger *slog.Logger
}

func NewDep(cfg *Config, logger *slog.Logger) *Dep {
	return &Dep{
		Cfg:    cfg,
		Logger: logger,
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
	return NewDep(cfg, logger), nil
}

func CloseDep(dep *Dep) {
	sentry.Flush(2 * time.Second)
}
